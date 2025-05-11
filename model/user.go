package model

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique" validate:"required,email"`
	Username  string    `json:"username" gorm:"unique" validate:"required,alphanum,min=3,max=20"`
	Rank      string    `json:"rank"`
	Password  string    `json:"password" validate:"required,min=6"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Rank     string `json:"rank"`
}
