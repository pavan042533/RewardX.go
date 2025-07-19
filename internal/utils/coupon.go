package utils

import (
	"fmt"
	"math/rand"
)

func GenerateCouponCode(prefix string) string{
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code:=make([]byte,6)
	for i:=range(code){
		code[i]=charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%v-%v", prefix,string(code))
}