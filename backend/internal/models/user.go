package models

import "time"

type User struct {
    ID        string    `db:"id" json:"id"`
    Email     string    `db:"email" json:"email"`
    Password  string    `db:"password" json:"-"`
    FullName  string    `db:"full_name" json:"full_name"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    FullName  string    `json:"full_name"`
    Accounts  []Account `json:"accounts"`
    CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}