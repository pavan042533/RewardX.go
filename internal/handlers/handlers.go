package handlers

import (
	"github.com/gofiber/fiber/v2"
	"authapi/internal/models"
	"authapi/internal/utils"
)

func RegisterUser(c *fiber.Ctx) error {
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if _, exists := models.Users[u.Username]; exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Username already exists"})
	}
	models.Users[u.Username] = u
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered"})
}

func LoginHandler(c *fiber.Ctx) error {
	var credentials models.User
	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, exists := models.Users[credentials.Username]
	if !exists || user.Password != credentials.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := utils.GenerateToken(credentials.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{"token": token})
}

func ShowPing(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "pong"})
}

func ViewProfile(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	return c.JSON(fiber.Map{"username": username})
}
