package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/app/model"
	"latihan-repository/app/repository"
	"latihan-repository/helper"
)

// PortfolioService memakai dua database: data mahasiswa dibaca dari
// PostgreSQL (apakah mahasiswanya ada, siapa pemiliknya), portofolionya
// disimpan di MongoDB. Tidak ada FOREIGN KEY lintas database, sehingga
// pemeriksaan "mahasiswanya ada" dilakukan di sini.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
	perms      *helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
	perms *helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, perms: perms}
}

// authorize memastikan mahasiswanya ada di PostgreSQL, lalu memeriksa hak
// akses: pemilik data student, atau pemegang permission anyPermission.
// Aturan yang sama dengan StudentService.
func (s *PortfolioService) authorize(c *fiber.Ctx, anyPermission string) (int, error) {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return 0, helper.BadRequest("id harus berupa angka positif")
	}

	current, ok := helper.CurrentUser(c)
	if !ok {
		return 0, helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.students.FindByID(ctx, studentID)
	if err != nil {
		return 0, translatePortfolioError(err)
	}

	if !CanAccessStudent(current, student.OwnerID, s.perms, anyPermission) {
		return 0, helper.Forbidden("tidak berhak mengakses portofolio mahasiswa lain")
	}

	return studentID, nil
}

// ---------- GET /students/:id/portfolio ----------
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:read:any")
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	p, err := s.portfolios.FindByStudent(ctx, studentID)
	if err != nil {
		return translateMongoPortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio ditemukan", p)
}

// ---------- PUT /students/:id/portfolio ----------
// Membuat (201) atau mengganti seluruh isi (200) portofolio.
func (s *PortfolioService) Replace(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any")
	if err != nil {
		return err
	}

	// Body didekode ke struct bertipe tetap, BUKAN map[string]any, sehingga
	// operator seperti {"$ne": ""} ditolak sebelum menyentuh MongoDB.
	var req model.ReplacePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req = normalizePortfolioRequest(req)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	p := model.Portfolio{
		StudentID: studentID, Headline: req.Headline, Bio: req.Bio,
	}
	for _, pr := range req.Projects {
		p.Projects = append(p.Projects, model.Project{
			Title: pr.Title, Technologies: pr.Technologies, URL: pr.URL, Year: pr.Year,
		})
	}
	for _, ce := range req.Certificates {
		p.Certificates = append(p.Certificates, model.Certificate{
			Name: ce.Name, Issuer: ce.Issuer, IssuedAt: ce.IssuedAt,
		})
	}

	saved, created, err := s.portfolios.Replace(ctx, p)
	if err != nil {
		return translateMongoPortfolioError(err)
	}

	if created {
		return helper.Created(c, "portofolio berhasil dibuat", saved,
			"/api/v1/students/"+c.Params("id")+"/portfolio")
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diganti", saved)
}

// ---------- POST /students/:id/portfolio/projects ----------
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any")
	if err != nil {
		return err
	}

	var req model.ProjectInput
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req = normalizeProject(req)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	updated, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title: req.Title, Technologies: req.Technologies, URL: req.URL, Year: req.Year,
	})
	if err != nil {
		return translateMongoPortfolioError(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", updated)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------
func (s *PortfolioService) RemoveProject(c *fiber.Ctx) error {
	studentID, err := s.authorize(c, "student:update:any")
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.portfolios.RemoveProject(ctx, studentID, c.Params("projectId")); err != nil {
		return translateMongoPortfolioError(err)
	}

	return helper.NoContent(c)
}

// ---------- GET /portfolios/stats/technologies ----------
// Statistik gabungan tanpa data pribadi, cukup untuk pengguna yang login.
func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return translateMongoPortfolioError(err)
	}

	return helper.Success(c, fiber.StatusOK, "10 teknologi terpopuler berhasil dihitung", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrPortfolioNotFound):
		return helper.NewError(fiber.StatusNotFound, "PORTFOLIO_NOT_FOUND",
			"portofolio belum dibuat, buat lebih dulu dengan PUT /students/:id/portfolio")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas 20")
	default:
		return helper.Internal(err)
	}
}

func translateMongoPortfolioError(err error) error {
	if repository.IsUnavailable(err) {
		return helper.ServiceUnavailable("MongoDB sementara tidak dapat dihubungi")
	}
	return translatePortfolioError(err)
}
