package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-repository/app/model"
	"latihan-repository/app/service"
	"latihan-repository/helper"
	"latihan-repository/middleware"
	"latihan-repository/route"
)

// merakit aplikasi: membuat instance Fiber, memasang middleware, lalu mendaftarkan route.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	prestasiService *service.PrestasiService,
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
	permissions *helper.PermissionSet,
) *fiber.App {

	app := fiber.New(fiber.Config{
		AppName: GetEnv(
			"APP_NAME",
			"Praktikum Backend Lanjut",
		),

		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger)

	route.Register(
		app,
		pool,
		studentService,
		prestasiService,
		authService,
		jwtManager,
		permissions,
	)

	// Penampung terakhir untuk URL yang tidak dikenal
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)

		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Kegagalan yang sudah kita rencanakan.

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}

		default:
			// Kegagalan yang tidak kita duga.
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status: fiberErr.Code, Code: "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", appErr.Error()))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}
