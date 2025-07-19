package models

import "time"
import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string    `json:"username"`
	Email    	 string    `gorm:"unique" json:"email"`
	Password 	 string    `json:"password"`
	Role     	 string    `json:"role"`
	Points   	 int       `json:"points"`
	IsVerified   bool      `json:"is_verified"`
	OTP          string    `json:"-"`
	OTPExpiresAt time.Time `json:"-"`
}
