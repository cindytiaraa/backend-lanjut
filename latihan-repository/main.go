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

	// =========================
	// Repository
	// =========================

	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	prestasiRepository := repository.NewPrestasiRepository(pool)
	prestasiService := service.NewPrestasiService(prestasiRepository)

	userRepository := repository.NewUserRepository(pool)
	refreshTokenRepository := repository.NewRefreshTokenRepository(pool)

	// =========================
	// JWT
	// =========================

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

	// =========================
	// Auth Service
	// =========================

	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		jwtManager,
		refreshTTL,
	)

	// =========================
	// Fiber App
	// =========================

	app := config.NewApp(
		logger,
		pool,
		studentService,
		prestasiService,
		authService,
		jwtManager,
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

	// =========================
	// Graceful Shutdown
	// =========================

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
