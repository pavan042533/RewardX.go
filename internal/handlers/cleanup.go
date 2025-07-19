package handlers

import (
	"time"
	"log"
	"authapi/internal/db"
	"authapi/internal/models"
)

// Delete users who never verified and are older than 30 minutes
func CleanUpUnverifiedUsers() {
	expiry := time.Now().Add(-30 * time.Minute)

	result := db.DB.
		Where("is_verified = ? AND created_at < ?", false, expiry).
		Delete(&models.User{})

	if result.Error != nil {
		log.Println("Cleanup failed:", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d unverified users\n", result.RowsAffected)
	}
}

