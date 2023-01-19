package models

import "time"

type Provider struct {
	ID           string `json:"id" gorm:"primaryKey"`
	ProviderName string `json:"provider_name" gorm:"not null"`
	User         *User
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type User struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	DisplayName  string    `json:"name" gorm:"not null"`
	Role         int       `json:"role" gorm:"not null"`
	Password     *string   `json:"-"`
	Email        string    `json:"email" gorm:"not null"`
	TokenVersion int       `json:"-" gorm:"default:1"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type AccountVerification struct {
	Email            string `gorm:"primaryKey"`
	VerificationCode string
	ExpiresAt        int64
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

type Email struct {
	Id         int `json:"id,omitempty" gorm:"primaryKey"`
	User       User
	UserID     string    `json:"user_id"`
	IsPrimary  bool      `json:"is_primary"`
	IsVerified bool      `json:"is_verified"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type EmailVerification struct {
	Id               int    `json:"id,omitempty" gorm:"primaryKey"`
	Email            Email  `gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE;"`
	EmailID          int    `json:"email_id"`
	VerificationCode string `json:"verification_code"`
	ExpiresAt        int64  `json:"expires_at"`
}

type Mask struct {
	Mask              string `json:"mask" gorm:"primaryKey"`
	Enabled           bool   `json:"enabled"`
	Email             Email  `gorm:"foreignKey:ForwardTo"`
	ForwardTo         int    `json:"forward_to"`
	User              User
	UserID            string    `json:"user_id"`
	MessagesReceived  int       `json:"messages_received" gorm:"default:0"`
	MessagesForwarded int       `json:"messages_forwarded" gorm:"default:0"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

type Domain struct {
	Domain    string    `json:"domain" gorm:"primaryKey"`
	Free      bool      `json:"free"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
