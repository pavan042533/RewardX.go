package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var SecurityKey = []byte("0425")

func GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(SecurityKey)
}

func ExtractSecretKey(token *jwt.Token) (interface{}, error) {
	return SecurityKey, nil
}
