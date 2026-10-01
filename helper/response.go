package helper

import "github.com/gofiber/fiber/v2"

// OK mengirim response 200 dengan data.
func OK(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// OKWithMeta mengirim response 200 dengan data + objek meta (untuk pagination).
func OKWithMeta(c *fiber.Ctx, message string, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// Created mengirim response 201 untuk resource yang baru dibuat.
func Created(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// NoContent mengirim response 204 (tidak ada body).
// Dipakai untuk DELETE yang berhasil.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
