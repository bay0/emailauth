package emailauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSecureCode(t *testing.T) {
	code, err := generateSecureCode()

	assert.NoError(t, err)
	assert.Len(t, code, codeLength)

	// Generate multiple codes and ensure they're different from each other
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := generateSecureCode()
		assert.NoError(t, err)
		assert.Len(t, code, codeLength)
		assert.False(t, codes[code], "Generated duplicate code")
		codes[code] = true
	}
}
