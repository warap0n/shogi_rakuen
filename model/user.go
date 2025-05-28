package model

import (
	"time"
)

// User はアプリのユーザーを表します
// Rate     : 初期値1500のレーティング
// RankType : "kyu" or "dan" を保持
// RankLevel: 級・段のレベル（デフォルト25=25級）
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Rate      int       `json:"rate" gorm:"default:1500;not null"`
	RankType  RankType  `json:"rank_type" gorm:"type:varchar(3);default:'kyu';not null"`
	RankLevel int       `json:"rank_level" gorm:"default:25;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserResponse struct {
	ID        uint     `json:"id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	RankType  RankType `json:"rank_type"`
	RankLevel int      `json:"rank_level"`
}
