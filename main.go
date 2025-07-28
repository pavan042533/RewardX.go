package main

import (
	"authapi/internal/handlers"
	"authapi/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"authapi/internal/db"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"os"
	"time"
)

func main() {
	godotenv.Load()
	db.Connect()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
    AllowOrigins:     "http://localhost:5173", 
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
    AllowCredentials: true,
}))

	// Public routes
	app.Post("/register", handlers.RegisterUser)
	app.Post("/verifyotp", handlers.VerifyOTP)
	app.Post("/login", handlers.LoginHandler)
	app.Get("/rewards", handlers.ListRewards)

	// user apis
	user := app.Group("/user",middleware.VerifyToken)
	user.Get("/profile", handlers.ViewProfile)
	user.Get("/wallet", handlers.GetUserWallet)
	user.Post("/redeem", handlers.RedeemReward)
	user.Get("/transactions", handlers.GetUserTransactions)

	// admin apis 
	admin := app.Group("/admin", middleware.VerifyToken)
	admin.Post("/addreward", handlers.AdminAddReward)
	admin.Post("/addpartner", handlers.AdminAddPartner)
	admin.Get("/getpartners", handlers.GetAllPartners)
	admin.Put("/rewards/:id", handlers.AdminUpdateReward)
	admin.Delete("/rewards/:id", handlers.AdminDeleteReward)
	admin.Get("/analytics", handlers.GetAdminAnalytics)

	// partner apis 
	partner := app.Group("/partner", middleware.VerifyToken)
	partner.Post("/addreward", handlers.PartnerAddReward)
	partner.Get("/rewards", handlers.GetPartnerRewards)
	partner.Put("/rewards/:id", handlers.PartnerUpdateReward)
	partner.Delete("/rewards/:id", handlers.PartnerDeleteReward)
	partner.Get("/analytics", handlers.GetPartnerAnalytics)

	port := os.Getenv("PORT")
	log.Fatal(app.Listen(":" + port))

	// Start background cleanup every 30 minutes
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			handlers.CleanUpUnverifiedUsers()
		}
	}()
}
