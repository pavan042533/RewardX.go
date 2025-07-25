package utils

import (
	"golang.org/x/crypto/bcrypt"
	// "golang.org/x/crypto v0.25.0"
)

func HashingPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	return string(bytes), err
}
func CheckPasswordHashing(pass, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}