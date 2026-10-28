package route

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"latihan-repository/app/model"
	"latihan-repository/app/service"
	"latihan-repository/helper"
	"latihan-repository/middleware"
)

// Register memetakan URL ke method pada service.
func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	mongoClient *mongo.Client,
	studentService *service.StudentService,
	prestasiService *service.PrestasiService,
	courseService *service.CourseService,
	supervisorService *service.SupervisorService,
	portfolioService *service.PortfolioService,
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
	api.Get("/health", healthCheck(pool, mongoClient))

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

	// PORTOFOLIO MAHASISWA (Modul 10, Tugas C). Data mahasiswa di PostgreSQL,
	// portofolio di MongoDB.
	students.Get("/:id/portfolio", portfolioService.Get)
	students.Put("/:id/portfolio", portfolioService.Replace)
	students.Post("/:id/portfolio/projects", portfolioService.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", portfolioService.RemoveProject)

	portfolios := api.Group(
		"/portfolios",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)
	portfolios.Get("/stats/technologies", portfolioService.TopTechnologies)

	// KELAS DAN KRS (Modul 9)
	courses := api.Group(
		"/courses",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", courseService.List)
	courses.Get("/mine", courseService.Mine)
	courses.Get("/:id", courseService.Get)
	courses.Post("/:id/enroll", courseService.Enroll)
	courses.Delete("/:id/enroll", courseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), courseService.UpdateQuota)

	enrollments := api.Group(
		"/enrollments",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)
	enrollments.Post("/swap", courseService.Swap)

	// PEMBIMBING TUGAS AKHIR (Tugas D.2)
	supervisors := api.Group(
		"/supervisors",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)
	supervisors.Get("/", supervisorService.List)
	supervisors.Get("/:id", supervisorService.Get)
	supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), supervisorService.Assign)
	supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), supervisorService.UpdateQuota)

	supervisions := api.Group(
		"/supervisions",
		middleware.RequireAuth(jwtManager),
		middleware.RequireJSON,
	)
	supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), supervisorService.Switch)

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

// HEALTH CHECK kini memeriksa DUA database. Aplikasi dianggap sehat hanya
// bila keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}

		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", fiber.Map{
			"timestamp": time.Now(),
		})
	}
}
