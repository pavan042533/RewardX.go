package models

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var Users = make(map[string]User)
