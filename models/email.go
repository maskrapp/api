package models

type Email struct {
	Id         int `json:"id,omitempty"`
	User       User
	UserID     string `json:"user_id"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
	Email      string `json:"email"`
}

type EmailVerification struct {
	Id               int    `json:"id,omitempty"`
	Email            Email  `gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE;"`
	EmailID          int    `json:"email_id"`
	VerificationCode string `json:"verification_code"`
	ExpiresAt        int64  `json:"expires_at"`
}

type Mask struct {
	Mask      string `json:"mask"`
	Enabled   bool   `json:"enabled"`
	Email     Email  `gorm:"foreignKey:ForwardTo"`
	ForwardTo int    `json:"forward_to"`
	User      User
	UserID    string `json:"user_id"`
}
