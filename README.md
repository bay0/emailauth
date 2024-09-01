# Email Authentication Package

This package provides a secure and flexible email authentication system using one-time codes. It's designed to be easily integrated into Go applications, offering both in-memory and Redis-based storage options for authentication codes.

## Features

- Secure code generation and verification
- Flexible storage options (in-memory and Redis)
- SMTP email sending capability
- Rate limiting to prevent abuse

## Installation

```bash
go get github.com/yourusername/emailauth
```

## Usage

Here's a basic example of how to use the package:

```go
import (
    "github.com/yourusername/emailauth"
)

// Initialize the email sender
emailSender := emailauth.NewSMTPEmailSender(
    "smtp.example.com",
    "587",
    "username",
    "password",
    "noreply@example.com",
    false,
)

// Initialize the code store (using in-memory store for this example)
codeStore := emailauth.NewInMemoryCodeStore()

// Create the auth service
authService := emailauth.NewAuthService(emailSender, codeStore)

// Send an authentication code
err := authService.SendAuthCode(ctx, "user@example.com")
if err != nil {
    // Handle error
}

// Verify the code
isValid, err := authService.VerifyCode(ctx, "user@example.com", "123456")
if err != nil {
    // Handle error
}
if isValid {
    // Code is valid, proceed with authentication
}
```

## Configuration

The package can be configured using environment variables. See the `example/main.go` file for a complete example of how to set up and run a server using this package.

## Testing

To run the tests:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
