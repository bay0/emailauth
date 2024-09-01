package emailauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCodeStore(t *testing.T) {
	store := NewInMemoryCodeStore()

	ctx := context.Background()
	email := "test@example.com"
	code := &AuthCode{
		Code:      "123456",
		ExpiresAt: time.Now().Add(100 * time.Millisecond), // shorter expiration for testing
	}

	// Test Set
	t.Run("Set Code", func(t *testing.T) {
		err := store.Set(ctx, email, code)
		assert.NoError(t, err)
	})

	// Test Get
	t.Run("Get Code", func(t *testing.T) {
		retrievedCode, err := store.Get(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedCode)
		assert.Equal(t, code.Code, retrievedCode.Code)
		assert.WithinDuration(t, code.ExpiresAt, retrievedCode.ExpiresAt, time.Millisecond)
	})

	// Test Expiration
	t.Run("Code Expiration", func(t *testing.T) {
		// Wait for the code to expire
		time.Sleep(150 * time.Millisecond)

		retrievedCode, err := store.Get(ctx, email)
		assert.NoError(t, err)
		assert.Nil(t, retrievedCode)
	})

	// Test Delete
	t.Run("Delete Code", func(t *testing.T) {
		// Reset the code
		err := store.Set(ctx, email, code)
		assert.NoError(t, err)

		err = store.Delete(ctx, email)
		assert.NoError(t, err)

		retrievedCode, err := store.Get(ctx, email)
		assert.NoError(t, err)
		assert.Nil(t, retrievedCode)
	})
}
