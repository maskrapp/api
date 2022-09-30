package auth

import "github.com/gofiber/fiber/v2"

func SignIn() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}
