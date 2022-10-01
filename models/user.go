package models

import "time"

type Provider struct {
	ProviderID   string `gorm:"primaryKey" json:"provider_id"`
	ProviderName string `json:"provider_name"`
	User         *User
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID            string  `json:"id" gorm:"primaryKey"`
	Name          string  `json:"name" gorm:"not null"`
	Role          int     `json:"role" gorm:"not null"`
	Password      *string `json:"-"`
	Email         string  `json:"email" gorm:"not null"`
	EmailVerified bool    `json:"email_verified" gorm:"not null"`
	Updated       int64   `gorm:"autoUpdateTime:milli"` // Use unix milli seconds as updating time
}
