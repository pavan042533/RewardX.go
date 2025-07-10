package main

import (
	"authapi/internal/handlers"
	"authapi/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Public routes
	app.Post("/register", handlers.RegisterUser)
	app.Post("/login", handlers.LoginHandler)
	app.Get("/ping", handlers.ShowPing)

	// Protected routes
	api := app.Group("/api", middleware.VerifyToken)
	api.Get("/profile", handlers.ViewProfile)

	app.Listen(":3000")
}
