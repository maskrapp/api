package models

import "time"

type User struct {
	ID                 string                    `json:"id"`
	Aud                string                    `json:"aud"`
	Role               string                    `json:"role"`
	Email              string                    `json:"email"`
	InvitedAt          time.Time                 `json:"invited_at"`
	ConfirmedAt        time.Time                 `json:"confirmed_at"`
	ConfirmationSentAt time.Time                 `json:"confirmation_sent_at"`
	AppMetadata        struct{ provider string } `json:"app_metadata"`
	UserMetadata       map[string]interface{}    `json:"user_metadata"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}
