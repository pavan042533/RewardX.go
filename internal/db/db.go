package db

import (
	"authapi/internal/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB = database
	DB.AutoMigrate(&models.User{}, &models.Reward{}, &models.Transaction{})
	SeedData()
}
func SeedData() {
	DB.Save(&models.User{
		Email:      "admin@rewardx.com",
		Username:   "Admin",
		Password:   "admin123",
		Role:       "admin",
		Points:      1000,
	    IsVerified:  true,
	})

	DB.Save(&models.User{
		Email:    "partner@brand.com",
		Username: "Partner",
		Password: "partner123",
		Role:     "partner",
		Points:   500,
		IsVerified:  true,
	})

	DB.Save(&models.Reward{Name: "Amazon Gift Card", Category: "Shopping", Cost: 100, Stock: 50})
	DB.Save(&models.Reward{Name: "Flipkart Voucher", Category: "Shopping", Cost: 80, Stock: 30})
	DB.Save(&models.Reward{Name: "Movie Tickets", Category: "Entertainment", Cost: 50, Stock: 20})

}