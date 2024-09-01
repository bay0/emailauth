package emailauth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/time/rate"
)

type AuthService struct {
	emailSender EmailSender
	store       CodeStore
	limiter     *rate.Limiter
}

func NewAuthService(emailSender EmailSender, store CodeStore) *AuthService {
	return &AuthService{
		emailSender: emailSender,
		store:       store,
		limiter:     rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests per second
	}
}

func (as *AuthService) SendAuthCode(ctx context.Context, email string) error {
	if err := as.limiter.Wait(ctx); err != nil {
		return errors.New("rate limit exceeded")
	}

	code, err := generateSecureCode()
	if err != nil {
		return err
	}

	err = as.emailSender.SendEmail(email, "Your Authentication Code", "Your authentication code is: "+code)
	if err != nil {
		return err
	}

	authCode := &AuthCode{
		Code:      code,
		ExpiresAt: time.Now().Add(codeExpiration),
	}

	return as.store.Set(ctx, email, authCode)
}

func (as *AuthService) VerifyCode(ctx context.Context, email, code string) (bool, error) {
	storedCode, err := as.store.Get(ctx, email)
	if err != nil {
		return false, err
	}

	if storedCode == nil {
		return false, errors.New("no code found for this email")
	}

	if time.Now().After(storedCode.ExpiresAt) {
		as.store.Delete(ctx, email)
		return false, errors.New("code has expired")
	}

	if storedCode.Code != code {
		return false, nil
	}

	as.store.Delete(ctx, email)
	return true, nil
}
