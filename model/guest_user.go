package model

import "time"

type GuestUser struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GuestUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}
