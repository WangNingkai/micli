package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// NotFound returns custom 404 page
func NotFound(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"code": 404,
		"msg":  "Not Found",
		"data": "",
	})
}
