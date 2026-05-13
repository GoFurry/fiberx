# mail addon

Reusable SMTP mail helper for Fiber-template projects.

This addon stays independent from `v3/*` templates so you can copy `mail.go` into any project and wire it at the application boundary.

## Features

- Multiple SMTP accounts in one service
- Pool rotation with `none`, `round_robin`, and `random`
- Automatic failover for retryable account/network/auth errors
- Custom HTML and plain-text content
- Built-in HTML templates: `welcome`, `verify_code`, `reset_password`
- Common mail fields: `cc`, `bcc`, `reply_to`, custom headers, attachments
- Standard-library only implementation

## Files

- `mail.go`: runtime implementation, designed to be copied as a single file
- `mail_test.go`: unit tests
- `go.mod`: standalone module boundary for local testing

## Quick Start

```go
package main

import (
	"context"
	"log"

	addonmail "github.com/gofurry/fiberx/addons/mail"
)

func main() {
	service, err := addonmail.New(addonmail.Config{
		Accounts: []addonmail.AccountConfig{
			{
				Name:       "primary",
				Host:       "smtp.example.com",
				Port:       587,
				Username:   "mailer@example.com",
				Password:   "secret",
				Encryption: addonmail.EncryptionSTARTTLS,
				From: addonmail.Address{
					Name:  "Example App",
					Email: "mailer@example.com",
				},
			},
			{
				Name:       "backup",
				Host:       "smtp-backup.example.com",
				Port:       587,
				Username:   "mailer-backup@example.com",
				Password:   "secret",
				Encryption: addonmail.EncryptionSTARTTLS,
				From: addonmail.Address{
					Name:  "Example App",
					Email: "mailer-backup@example.com",
				},
			},
		},
		EnableRotation:   true,
		RotationStrategy: addonmail.RotationStrategyRoundRobin,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = service.Send(context.Background(), addonmail.Message{
		To:       []string{"user@example.com"},
		Subject:  "Welcome",
		TextBody: "Welcome to Example App",
		HTMLBody: "<h1>Welcome to Example App</h1>",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Built-in Templates

### `welcome`

Suggested `Data` keys:

- `app_name`
- `headline`
- `recipient_name`
- `intro`
- `action_url`
- `action_text`
- `footer`

### `verify_code`

Suggested `Data` keys:

- `app_name`
- `recipient_name`
- `code`
- `expires_in`
- `footer`

### `reset_password`

Suggested `Data` keys:

- `app_name`
- `recipient_name`
- `reset_url`
- `action_text`
- `expires_in`
- `footer`

Example:

```go
err = service.SendTemplate(context.Background(), addonmail.TemplateMessage{
	Message: addonmail.Message{
		To:      []string{"user@example.com"},
		Subject: "Reset your password",
	},
	Template: addonmail.TemplateResetPassword,
	Data: map[string]any{
		"app_name":       "Example App",
		"recipient_name": "Alice",
		"reset_url":      "https://example.com/reset?token=abc",
		"expires_in":     "30 minutes",
	},
})
```

## Notes

- `EnableRotation=false` pins the initial account selection to the first account, but retryable failures still fail over to the next account.
- `Bcc` is added to the SMTP envelope only and will not be written into the MIME headers.
- `HTMLBody` and `SendTemplate(...)` should not be mixed in the same message.
- Attachments support either in-memory `Data` or loading from `Path`.
- Empty `AccountConfig.From` falls back to `Config.DefaultFrom`, then to the account username if it is a valid email address.

## Local Test

```bash
cd addons/mail
go test ./...
```
