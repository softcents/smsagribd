package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// databaseError returns a 500 JSON error response for unexpected DB errors.
func databaseError(ctx fiber.Ctx, err error) error {
	return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"message": "A database error occurred.",
		"error":   err.Error(),
	})
}
