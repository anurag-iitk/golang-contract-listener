package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse sends an error response to the client
func ErrorResponse(c *fiber.Ctx, status int, message string, err error) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   message,
		"details": err.Error(),
	})
}
