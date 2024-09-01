package emailauth

import (
	"context"
	"time"
)

type CodeStore interface {
	Set(ctx context.Context, email string, code *AuthCode) error
	Get(ctx context.Context, email string) (*AuthCode, error)
	Delete(ctx context.Context, email string) error
}

type EmailSender interface {
	SendEmail(to, subject, body string) error
}

type AuthCode struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}
