package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/app/model"
	"latihan-repository/app/repository"
	"latihan-repository/helper"
)

// SupervisorService menangani penetapan dosen pembimbing tugas akhir untuk
// entity students. Pembagian tugasnya sama dengan CourseService:
//   - method yang menerima *fiber.Ctx hanya membaca permintaan, memanggil
//     use case, lalu menyusun respons (peran controller);
//   - method huruf kecil (assign, switchSupervisor) menjalankan alur kerja
//     di dalam transaction dan TIDAK mengenal Fiber (peran use case);
//   - aturan murninya ada di supervisor_rules.go.
type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

// NewSupervisorService menerima INTERFACE repository dan TxManager.
func NewSupervisorService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

// errSameSupervisor: pergantian ke pembimbing yang sama tidak bermakna.
var errSameSupervisor = errors.New("pembimbing baru sama dengan pembimbing lama")

// ---------- GET /supervisors ----------
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	sup, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", sup)
}

// ---------- POST /supervisors/:id/assign ----------
// Dijaga permission supervision:manage.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	assigned, err := s.assign(ctx, supervisorID, req.StudentID)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "pembimbing berhasil ditetapkan", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// assign adalah use case penetapan pembimbing. Tidak mengenal Fiber.
func (s *SupervisorService) assign(ctx context.Context, supervisorID, studentID int) (model.Supervisor, error) {
	var assigned model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci baris student (pessimistic). Dua request untuk mahasiswa
		//    yang SAMA kini antre; yang kedua akan melihat hasil yang
		//    pertama dan ditolak oleh constraint unik pada langkah 3.
		if err := repo.LockStudent(ctx, studentID); err != nil {
			return err
		}

		// 2. Ambil satu slot bimbingan secara atomik dan bersyarat.
		var err error
		assigned, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		// 3. Catat penetapannya. Bila mahasiswa sudah punya pembimbing,
		//    error di sini membatalkan juga langkah 2 lewat rollback.
		return repo.InsertSupervision(ctx, studentID, supervisorID)
	})

	return assigned, err
}

// ---------- POST /supervisions/switch ----------
// Dijaga permission supervision:manage.
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisionRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	joined, err := s.switchSupervisor(ctx, req)
	if err != nil {
		if errors.Is(err, errSameSupervisor) {
			return helper.Validation(map[string]string{
				"to_supervisor_id": "tidak boleh sama dengan pembimbing saat ini",
			})
		}
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "pembimbing berhasil diganti", joined)
}

// switchSupervisor adalah use case pergantian pembimbing. Tidak mengenal Fiber.
//
// Atomik: bila pembimbing baru penuh, mahasiswa TETAP dengan pembimbing
// lama. Bebas deadlock: kedua pembimbing dikunci dalam urutan id menaik.
func (s *SupervisorService) switchSupervisor(ctx context.Context, req model.SwitchSupervisionRequest) (model.Supervisor, error) {
	var joined model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci student DULU. Tanpa kunci ini, dua pergantian serentak
		//    untuk mahasiswa yang sama sama-sama membaca pembimbing lama
		//    "A", lalu sama-sama mengurangi A dan menambah tujuannya
		//    masing-masing: penghitung menyimpang dari data penetapan.
		//    Pengambilan slot yang bersyarat hanya menjaga KUOTA, bukan
		//    kebenaran "pembimbing lama" yang dibaca di Go.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Baca pembimbing lama SETELAH kunci dipegang (nilainya pasti
		//    terbaru).
		fromID, err := repo.FindSupervisorOf(ctx, req.StudentID)
		if err != nil {
			return err
		}
		if IsSameSupervisor(fromID, req.ToSupervisorID) {
			return errSameSupervisor
		}

		// 3. Kunci kedua pembimbing dalam urutan id menaik (anti-deadlock).
		if err := repo.LockSupervisors(ctx, fromID, req.ToSupervisorID); err != nil {
			return err
		}

		// 4. Lepas dari yang lama, ambil slot di yang baru. Bila yang baru
		//    penuh, error ini me-rollback pengurangan di yang lama juga.
		if _, err := repo.DecrementSupervised(ctx, fromID); err != nil {
			return err
		}
		joined, err = repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}
		return repo.MoveSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})

	return joined, err
}

// ---------- PATCH /supervisors/:id ----------
// Optimistic locking. Dijaga permission supervision:manage.
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

// translateSupervisorError memetakan error milik repository menjadi AppError.
func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki pembimbing")
	case errors.Is(err, repository.ErrSupervisorVersion):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrSupervisorQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}
