package emailauth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func TestRedisCodeStore(t *testing.T) {
	mockRedis := new(MockRedisClient)
	store := NewRedisCodeStore(mockRedis)

	ctx := context.Background()
	email := "test@example.com"
	code := &AuthCode{
		Code:      "123456",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Test Set
	encodedCode, _ := json.Marshal(code)
	mockRedis.On("Set", ctx, email, string(encodedCode), codeExpiration).Return(nil)

	err := store.Set(ctx, email, code)
	assert.NoError(t, err)

	// Test Get
	mockRedis.On("Get", ctx, email).Return(string(encodedCode), nil)

	retrievedCode, err := store.Get(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, code.Code, retrievedCode.Code)
	assert.WithinDuration(t, code.ExpiresAt, retrievedCode.ExpiresAt, time.Second)

	// Test Delete
	mockRedis.On("Del", ctx, []string{email}).Return(nil)

	err = store.Delete(ctx, email)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
}
