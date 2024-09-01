package emailauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSMTPEmailSender(t *testing.T) {
	sender := NewSMTPEmailSender("smtp.example.com", "587", "username", "password", "noreply@example.com", true)

	assert.NotNil(t, sender)
	assert.Equal(t, "smtp.example.com", sender.smtpHost)
	assert.Equal(t, "587", sender.smtpPort)
	assert.Equal(t, "username", sender.smtpUsername)
	assert.Equal(t, "password", sender.smtpPassword)
	assert.Equal(t, "noreply@example.com", sender.senderEmail)
	assert.True(t, sender.allowUnencrypted)
}
