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
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
) {
	// Endpoint root dari project sebelumnya.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck(pool))

	// AUTH
	auth := api.Group("/auth")

	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimit(), authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)

	auth.Get(
		"/me",
		middleware.RequireAuth(jwtManager),
		authService.Me,
	)

	// STUDENT
	students := api.Group(
		"/students",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)

	// RBAC Student berbasis permission
	students.Get("/", middleware.RequirePermission(perms, "student:list"), studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), studentService.Delete)

	// PRESTASI
	prestasi := api.Group(
		"/prestasi",
		middleware.RequireJSON,
	)

	prestasi.Get("/", prestasiList(prestasiService))
	prestasi.Get("/:id", prestasiGet(prestasiService))
	prestasi.Post("/", prestasiCreate(prestasiService))
	prestasi.Delete("/:id", prestasiDelete(prestasiService))
}

// ==========================
// PRESTASI HANDLER
// ==========================
func prestasiList(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, err := s.List(ctx)
		if err != nil {
			return helper.Internal(err)
		}

		return helper.Success(c, fiber.StatusOK, "data prestasi berhasil diambil", data)
	}
}

func prestasiGet(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := helper.ParamID(c)
		if !ok {
			return helper.BadRequest("id tidak valid")
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, err := s.Get(ctx, id)
		if err != nil {
			if service.IsPrestasiNotFound(err) {
				return helper.NotFound("prestasi tidak ditemukan")
			}
			return helper.Internal(err)
		}

		return helper.Success(c, fiber.StatusOK, "data prestasi berhasil diambil", data)
	}
}

func prestasiCreate(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.CreatePrestasiRequest

		if err := c.BodyParser(&req); err != nil {
			return helper.BadRequest("body tidak valid")
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		data, errs, err := s.Create(ctx, req)

		if len(errs) > 0 {
			return helper.Validation(errs)
		}

		if err != nil {
			return helper.Internal(err)
		}

		return helper.Created(c, "prestasi berhasil ditambahkan", data, "/api/v1/prestasi/"+strconv.Itoa(data.IDPrestasi))
	}
}

func prestasiDelete(s *service.PrestasiService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := helper.ParamID(c)
		if !ok {
			return helper.BadRequest("id tidak valid")
		}

		ctx, cancel := helper.RequestContext(c)
		defer cancel()

		err := s.Delete(ctx, id)

		if err != nil {
			if service.IsPrestasiNotFound(err) {
				return helper.NotFound("prestasi tidak ditemukan")
			}
			return helper.Internal(err)
		}

		return helper.NoContent(c)
	}
}

// HEALTH CHECK
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", fiber.Map{
			"timestamp": time.Now(),
		})
	}
}
