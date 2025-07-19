package utils

import(
	"math/rand"
	// "time"
	"strconv"
)

func GenerateOTP() string {
	// rand.Seed(time.Now().UnixNano())
    otp := rand.Intn(900000) + 100000 // 6-digit
    return strconv.Itoa(otp)
}