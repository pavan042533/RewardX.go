package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "os"
)

var SecurityKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(id uint, role string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": id,
        "role": role,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })
    return token.SignedString(SecurityKey)
}

func ExtractSecretKey(token *jwt.Token) (interface{}, error) {
    return SecurityKey,nil
}