package emailauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCodeStore struct {
	mock.Mock
}

func (m *MockCodeStore) Set(ctx context.Context, email string, code *AuthCode) error {
	args := m.Called(ctx, email, code)
	return args.Error(0)
}

func (m *MockCodeStore) Get(ctx context.Context, email string) (*AuthCode, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*AuthCode), args.Error(1)
}

func (m *MockCodeStore) Delete(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestSendAuthCode(t *testing.T) {
	mockStore := new(MockCodeStore)
	mockEmailSender := new(MockEmailSender)

	authService := NewAuthService(mockEmailSender, mockStore)

	ctx := context.Background()
	email := "test@example.com"

	mockEmailSender.On("SendEmail", email, mock.Anything, mock.Anything).Return(nil)
	mockStore.On("Set", ctx, email, mock.AnythingOfType("*emailauth.AuthCode")).Return(nil)

	err := authService.SendAuthCode(ctx, email)

	assert.NoError(t, err)
	mockEmailSender.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}

func TestVerifyCode(t *testing.T) {
	mockStore := new(MockCodeStore)
	mockEmailSender := new(MockEmailSender)

	authService := NewAuthService(mockEmailSender, mockStore)

	ctx := context.Background()
	email := "test@example.com"
	code := "123456"

	// Test valid code and expiration date
	mockStore.On("Get", ctx, email).Return(&AuthCode{
		Code:      code,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil).Once()
	mockStore.On("Delete", ctx, email).Return(nil).Once()

	isValid, err := authService.VerifyCode(ctx, email, code)

	assert.NoError(t, err)
	assert.True(t, isValid)

	// Test invalid code and expiration date
	mockStore.On("Get", ctx, email).Return(&AuthCode{
		Code:      "654321",
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil).Once()

	isValid, err = authService.VerifyCode(ctx, email, code)

	assert.NoError(t, err)
	assert.False(t, isValid)

	mockStore.AssertExpectations(t)
}
