package models

type EmailEntry struct {
	Id         int    `json:"id,omitempty"`
	UserId     string `json:"user_id"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
	Email      string `json:"email"`
}

type EmailVerificationEntry struct {
	Id               int    `json:"id,omitempty"`
	EmailId          int    `json:"email_id"`
	VerificationCode string `json:"verification_code"`
	ExpiresAt        int64  `json:"expires_at"`
}

type MaskEntry struct {
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	ForwardTo int    `json:"forward_to"`
	UserId    string `json:"user_id"`
}
