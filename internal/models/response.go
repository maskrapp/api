package models

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
