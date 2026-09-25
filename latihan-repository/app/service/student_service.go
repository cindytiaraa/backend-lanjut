package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/app/model"
	"latihan-repository/app/repository"
	"latihan-repository/helper"
)

// menerima *fiber.Ctx (peran controller) dan menjalankan business rules (peran use case).
type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

// menerima INTERFACE, bukan struct konkret.
func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// List melayani GET /students dengan cursor pagination DAN dua format (JSON/CSV) lewat content negotiation.
func (s *StudentService) List(c *fiber.Ctx) error {
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data student lain")
	}

	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	baru, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  current.UserID,
	})
	if err != nil {
		return translateError(err, "student")
	}

	return helper.Created(c, "student berhasil dibuat", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
		OwnerID:  student.OwnerID,
	})
	if err != nil {
		return translateError(err, "student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", hasil)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// PATCH = baca dulu, ubah seperlunya, lalu simpan kembali.
	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "student")
	}

	if !CanAccessStudent(current, saatIni.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student lain")
	}

	updated := ApplyPatch(saatIni, req)

	hasil, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "student")
	}

	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", hasil)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, "student")
	}

	return helper.NoContent(c)
}

func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("NIM sudah dipakai")
	default:
		return helper.Internal(err)
	}
}
