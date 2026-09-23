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

	// Repository
	studentRepository := repository.NewStudentRepository(pool)
	prestasiRepository := repository.NewPrestasiRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	refreshTokenRepository := repository.NewRefreshTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentService := service.NewStudentService(studentRepository, permissions)
	prestasiService := service.NewPrestasiService(prestasiRepository)

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
		studentService,
		prestasiService,
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
