package helper

import (
	"github.com/gofiber/fiber/v2"

	"latihan-repository/app/model"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessCursor mengirim daftar data beserta CursorMeta (next_cursor, has_more).
func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim 201 sekaligus memasang header Location.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Catatan: helper.Fail dan helper.FailValidation SENGAJA dihapus.
// Sejak Modul 7, handler tidak lagi menulis response kegagalan sendiri —
// ia hanya mengembalikan *helper.AppError (lihat helper/errors.go), dan
// satu-satunya yang mengubahnya menjadi response HTTP adalah ErrorHandler
// terpusat di config/app.go. Menghapus fungsi lama ini membuat compiler
// yang menegakkan aturan itu, bukan disiplin kita sendiri.
