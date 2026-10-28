package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"latihan-repository/app/repository"
	"latihan-repository/app/service"
	"latihan-repository/config"
	"latihan-repository/database"
	"latihan-repository/helper"
)

func main() {
	config.LoadEnv()

	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal membuat database pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// Disconnect menutup seluruh koneksi di pool milik driver. Diberi batas
	// waktu sendiri, karena context milik request sudah tidak berlaku saat
	// aplikasi berhenti.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	// Repository
	studentRepository := repository.NewStudentRepository(pool)
	prestasiRepository := repository.NewPrestasiRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	refreshTokenRepository := repository.NewRefreshTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	courseRepository := repository.NewCourseRepository(pool)
	supervisorRepository := repository.NewSupervisorRepository(pool)

	portfolioRepository := repository.NewPortfolioRepository(mongoDB)
	if err := portfolioRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index MongoDB",
			slog.String("collection", "student_portfolios"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	reviewRepository := repository.NewReviewRepository(mongoDB)
	if err := reviewRepository.EnsureIndexes(context.Background()); err != nil {
		logger.Error("gagal membuat index MongoDB",
			slog.String("collection", "course_reviews"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	// TxManager: satu-satunya pintu untuk membuka dan menutup transaction.
	txManager := repository.NewTxManager(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentService := service.NewStudentService(studentRepository, permissions)
	prestasiService := service.NewPrestasiService(prestasiRepository)
	courseService := service.NewCourseService(courseRepository, txManager)
	supervisorService := service.NewSupervisorService(supervisorRepository, txManager)
	portfolioService := service.NewPortfolioService(portfolioRepository, studentRepository, permissions)
	reviewService := service.NewReviewService(reviewRepository, courseRepository, permissions)

	// JWT
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Error("JWT_SECRET belum dikonfigurasi")
		os.Exit(1)
	}

	jwtIssuer := config.GetEnv(
		"JWT_ISSUER",
		"praktikum-backend",
	)

	accessTTL := time.Duration(
		config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15),
	) * time.Minute

	refreshTTL := time.Duration(
		config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7),
	) * 24 * time.Hour

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		jwtIssuer,
		accessTTL,
	)

	// Auth Service
	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		jwtManager,
		permissions,
		refreshTTL,
	)

	// Fiber App
	app := config.NewApp(
		logger,
		pool,
		mongoClient,
		studentService,
		prestasiService,
		courseService,
		supervisorService,
		portfolioService,
		reviewService,
		authService,
		jwtManager,
		permissions,
	)

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error(
				"gagal menjalankan server",
				slog.Any("error", err),
			)
		}
	}()

	logger.Info(
		"server berjalan",
		slog.String("port", port),
	)

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	logger.Info("mematikan server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := app.Shutdown(); err != nil {
		logger.Error(
			"gagal shutdown server",
			slog.Any("error", err),
		)
	}

	<-shutdownCtx.Done()

	logger.Info("server berhenti")
}
