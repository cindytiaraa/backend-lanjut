package route

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-repository/app/model"
	"latihan-repository/app/service"
	"latihan-repository/helper"
	"latihan-repository/middleware"
)

// Register memetakan URL ke method pada service.
func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	prestasiService *service.PrestasiService,
) {
	// Endpoint root dari project sebelumnya.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck(pool))

	// =========================
	// STUDENT
	// =========================

	students := api.Group(
		"/students",
		middleware.RequireJSON,
	)

	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)

	// =========================
	// PRESTASI
	// =========================

	prestasi := api.Group(
		"/prestasi",
		middleware.RequireJSON,
	)

	prestasi.Get("/", prestasiList(prestasiService))
	prestasi.Get("/:id", prestasiGet(prestasiService))
	prestasi.Post("/", prestasiCreate(prestasiService))
	prestasi.Delete("/:id", prestasiDelete(prestasiService))
}

// =========================
// PRESTASI HANDLER
// =========================

func prestasiList(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, err := s.List(ctx)
		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil data prestasi",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"data prestasi berhasil diambil",
			data,
		)
	}
}

func prestasiGet(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := helper.ParamID(c)
		if !ok {
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id tidak valid",
			)
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, err := s.Get(ctx, id)
		if err != nil {
			if service.IsPrestasiNotFound(err) {
				return helper.Fail(
					c,
					fiber.StatusNotFound,
					"prestasi tidak ditemukan",
				)
			}

			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal mengambil data prestasi",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"data prestasi berhasil diambil",
			data,
		)
	}
}

func prestasiCreate(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreatePrestasiRequest

		if err := c.BodyParser(&req); err != nil {
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"body tidak valid",
			)
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, errs, err := s.Create(ctx, req)

		if len(errs) > 0 {
			return helper.FailValidation(c, errs)
		}

		if err != nil {
			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal menambahkan prestasi",
			)
		}

		return helper.Created(
			c,
			"prestasi berhasil ditambahkan",
			data,
			"/api/v1/prestasi/"+strconv.Itoa(data.IDPrestasi),
		)
	}
}

func prestasiDelete(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := helper.ParamID(c)
		if !ok {
			return helper.Fail(
				c,
				fiber.StatusBadRequest,
				"id tidak valid",
			)
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		err := s.Delete(ctx, id)

		if err != nil {
			if service.IsPrestasiNotFound(err) {
				return helper.Fail(
					c,
					fiber.StatusNotFound,
					"prestasi tidak ditemukan",
				)
			}

			return helper.Fail(
				c,
				fiber.StatusInternalServerError,
				"gagal menghapus prestasi",
			)
		}

		return helper.NoContent(c)
	}
}

// =========================
// HEALTH CHECK
// =========================

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"server dan database berjalan",
			fiber.Map{
				"timestamp": time.Now(),
			},
		)
	}
}
