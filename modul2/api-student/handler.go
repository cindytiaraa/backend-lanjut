package main

// Bagian 1: penyimpanan dan bantuan
import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var users []Student
var nextID = 1

func init() {
	users = []Student{
		{
			ID:        1,
			NIM:       "20260001",
			Name:      "Cindy",
			Grade:     90,
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}
	nextID = 2
}

func findStudentIndex(id int) int {
	for i := range users {
		if users[i].ID == id {
			return i
		}
	}
	return -1
}

// cocokPencarian memeriksa apakah kata kunci muncul di NIM atau Name.
func cocokPencarian(u Student, kata string) bool {
	kata = strings.ToLower(kata)
	return strings.Contains(strings.ToLower(u.NIM), kata) ||
		strings.Contains(strings.ToLower(u.Name), kata)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// Bagian 2: daftar dengan saring, urut, dan penggal
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	// 1) Saring
	hasil := []Student{}
	for _, u := range users {
		if q.IsActive != nil && u.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" && !cocokPencarian(u, q.Search) {
			continue
		}
		hasil = append(hasil, u)
	}

	// 2) Urutkan
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name":
			lebihKecil = hasil[i].Name < hasil[j].Name
		case "created_at":
			lebihKecil = hasil[i].CreatedAt.Before(hasil[j].CreatedAt)
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// 3) Potong sesuai halaman
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar student berhasil diambil", hasil[mulai:akhir], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

// Bagian 3: ambil satu dan tambah baru
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	return ok(c, "student ditemukan", users[i])
}

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0-100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	for _, u := range users {
		if strings.EqualFold(u.NIM, req.NIM) {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai")
		}
	}

	baru := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	users = append(users, baru)
	nextID++

	return created(c, "student berhasil dibuat", baru,

		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

// Bagian 4: mengganti dan menghapus student
// PUT mengganti SELURUH isi. Field yang tidak dikirim dianggap dikosongkan,
// karena itu semuanya wajib ada.
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0-100 pada PUT"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	for _, u := range users {
		if u.ID != users[i].ID && strings.EqualFold(u.NIM, req.NIM) {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai")
		}
	}

	users[i].NIM = req.NIM
	users[i].Name = req.Name
	users[i].Grade = req.Grade
	users[i].IsActive = req.IsActive

	return ok(c, "student berhasil diganti seluruhnya", users[i])
}

// PATCH hanya mengubah field yang benar-benar dikirim.
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		users[i].NIM = *req.NIM
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		users[i].Name = *req.Name
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{"grade": "harus antara 0-100"})
		}
		users[i].Grade = *req.Grade
	}
	if req.IsActive != nil {
		users[i].IsActive = *req.IsActive
	}

	return ok(c, "student berhasil diperbarui sebagian", users[i])
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	users = append(users[:i], users[i+1:]...)

	return noContent(c) // 204: berhasil, dan memang tidak ada yang perlu dikirim
}
