package emailauth

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCodeStore(t *testing.T) {
	store := NewInMemoryCodeStore()

	ctx := context.Background()
	email := "test@example.com"
	code := &AuthCode{
		Code:      "123456",
		ExpiresAt: time.Now().Add(time.Minute), // shorter expiration for testing
	}

	// Test Set
	t.Run("Set Code", func(t *testing.T) {
		err := store.Set(ctx, email, code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Test Get
	t.Run("Get Code", func(t *testing.T) {
		retrievedCode, err := store.Get(ctx, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if retrievedCode == nil {
			t.Fatalf("expected code, got nil")
		}
		if retrievedCode.Code != code.Code {
			t.Errorf("expected code %s, got %s", code.Code, retrievedCode.Code)
		}
		if !retrievedCode.ExpiresAt.Equal(code.ExpiresAt) {
			t.Errorf("expected expiration date %v, got %v", code.ExpiresAt, retrievedCode.ExpiresAt)
		}
	})

	// Test Expiration
	t.Run("Code Expiration", func(t *testing.T) {
		// Wait for the code to expire
		time.Sleep(time.Minute + time.Second)

		retrievedCode, err := store.Get(ctx, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if retrievedCode != nil {
			t.Fatalf("expected code to be expired and deleted, but it still exists")
		}
	})

	// Test Delete
	t.Run("Delete Code", func(t *testing.T) {
		// Reset the code
		err := store.Set(ctx, email, code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = store.Delete(ctx, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrievedCode, err := store.Get(ctx, email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if retrievedCode != nil {
			t.Errorf("expected code to be deleted, but it still exists")
		}
	})
}
