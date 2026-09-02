package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/app/model"
	"latihan-repository/app/repository"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

// Perhatikan tipe parameternya: INTERFACE, bukan struct konkret.
// Handler tidak tahu dan tidak perlu tahu datanya disimpan di mana.
func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// terjemahkanError memetakan error milik repository menjadi status HTTP.
func terjemahkanError(c *fiber.Ctx, err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai")
	default:
		return fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := parseListQuery(c)

	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return okList(c, "daftar student berhasil diambil", students, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data student")
	}

	return ok(c, "student ditemukan", student)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
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

	// Keunikan NIM TIDAK diperiksa dengan SELECT lebih dulu.
	// Basis data sudah menjaminnya lewat UNIQUE INDEX pada LOWER(nim),
	// dan pemeriksaan manual justru menyisakan race condition.
	baru, err := h.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal menyimpan student")
	}

	return created(c, "student berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
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

	hasil, err := h.repo.Update(ctx, model.Student{
		ID: id, NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
	})
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui student")
	}

	return ok(c, "student berhasil diganti seluruhnya", hasil)
}

func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	// PATCH = baca dulu, ubah seperlunya, lalu simpan kembali.
	saatIni, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c, err, "gagal mengambil data student")
	}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		saatIni.NIM = *req.NIM
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		saatIni.Name = *req.Name
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{"grade": "harus antara 0-100"})
		}
		saatIni.Grade = *req.Grade
	}
	if req.IsActive != nil {
		saatIni.IsActive = *req.IsActive
	}

	hasil, err := h.repo.Update(ctx, saatIni)
	if err != nil {
		return terjemahkanError(c, err, "gagal memperbarui student")
	}

	return ok(c, "student berhasil diperbarui sebagian", hasil)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return terjemahkanError(c, err, "gagal menghapus student")
	}

	return noContent(c)
}
