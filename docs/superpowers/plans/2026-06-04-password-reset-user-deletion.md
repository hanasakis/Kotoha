# Password Reset & User Deletion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add password reset (Resend email) and soft user deletion to the Kotoha auth module.

**Architecture:** New `pkg/mail` SMTP package wraps Resend. `PasswordResetToken` model stores SHA256(token) with 1h expiry. Auth service gains `ForgotPassword`, `ResetPassword`, `DeleteAccount` methods. Routes: `POST /auth/forgot-password`, `POST /auth/reset-password`, `DELETE /auth/account`.

**Tech Stack:** `net/smtp` (stdlib), `crypto/sha256` (stdlib), `crypto/rand` (stdlib), GORM, bcrypt

---

## File Map

| Action | File | Responsibility |
|--------|------|---------------|
| Create | `pkg/mail/mail.go` | SMTP client (Resend) |
| Create | `internal/auth/password_reset.go` | Token generation, email template, orchestration |
| Modify | `internal/auth/models.go` | Add `PasswordResetToken` struct |
| Modify | `internal/auth/repository.go` | CRUD for `PasswordResetToken` |
| Modify | `internal/auth/service.go` | `ForgotPassword`, `ResetPassword`, `DeleteAccount` |
| Modify | `internal/auth/handler.go` | 3 new handler methods |
| Modify | `internal/config/config.go` | Add `SMTPConfig` struct + env vars |
| Modify | `internal/router/router.go` | Create mail client, pass to auth, register routes |
| Modify | `internal/i18n/locales/zh.json` | 3 new keys |
| Modify | `internal/i18n/locales/en.json` | 3 new keys |

---

### Task 1: SMTP Mail Package

**Files:**
- Create: `pkg/mail/mail.go`

- [ ] **Step 1: Create the mail package**

```go
package mail

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Client struct {
	host     string
	port     string
	username string
	password string
	from     string
	fromName string
}

func New(host, port, username, password, from, fromName string) *Client {
	return &Client{host: host, port: port, username: username, password: password, from: from, fromName: fromName}
}

func (c *Client) Send(to, subject, htmlBody string) error {
	addr := c.host + ":" + c.port
	auth := smtp.PlainAuth("", c.username, c.password, c.host)

	headers := []string{
		"From: " + c.fromName + " <" + c.from + ">",
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody

	return smtp.SendMail(addr, auth, c.from, []string{to}, []byte(msg))
}

func (c *Client) SendPasswordReset(to, token string) error {
	subject := "Kotoha Password Reset"
	resetURL := "https://kotoha.shop/reset-password?token=" + token
	body := `<html><body>
<p>You requested a password reset for your Kotoha account.</p>
<p>Click the link below to reset your password (valid for 1 hour):</p>
<p><a href="` + resetURL + `">` + resetURL + `</a></p>
<p>If you did not request this, please ignore this email.</p>
</body></html>`
	return c.Send(to, subject, body)
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./pkg/mail/...`
Expected: success

---

### Task 2: SMTP Config

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add SMTPConfig struct and field to Config**

Add after `CORSConfig` (around line 74):

```go
type SMTPConfig struct {
	Host, Port, User, Password, From, FromName string
}
```

Add to `Config` struct after `CORS CORSConfig`:

```go
SMTP SMTPConfig
```

- [ ] **Step 2: Add SMTPConfig to Load()**

Add to the config literal in `Load()` (before the closing `}`):

```go
SMTP: SMTPConfig{
	Host:     getEnv("SMTP_HOST", "smtp.resend.com"),
	Port:     getEnv("SMTP_PORT", "587"),
	User:     getEnv("SMTP_USER", "resend"),
	Password: getEnv("SMTP_PASSWORD", ""),
	From:     getEnv("SMTP_FROM", "onboarding@resend.dev"),
	FromName: getEnv("SMTP_FROM_NAME", "Kotoha"),
},
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 3: PasswordResetToken Model

**Files:**
- Modify: `internal/auth/models.go`

- [ ] **Step 1: Add PasswordResetToken struct**

Add after the `Session` struct in `internal/auth/models.go`:

```go
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null" json:"-"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 2: Add to auto-migration**

Read `pkg/db/migrate.go` to check if models are listed. If auto-migrate uses a list, add `&auth.PasswordResetToken{}`.

Run grep: `grep -r "AutoMigrate" pkg/db/`

Expected: Find the auto-migrate call and add the new model.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 4: Password Reset Repository

**Files:**
- Modify: `internal/auth/repository.go`

- [ ] **Step 1: Add reset token CRUD methods**

Append to `internal/auth/repository.go`:

```go
func (r *Repository) CreatePasswordResetToken(token *PasswordResetToken) error {
	return r.DB.Create(token).Error
}

func (r *Repository) FindPasswordResetToken(tokenHash string) (*PasswordResetToken, error) {
	var t PasswordResetToken
	err := r.DB.Where("token_hash = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) MarkResetTokenUsed(tokenID uint) error {
	return r.DB.Model(&PasswordResetToken{}).Where("id = ?", tokenID).Update("used", true).Error
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 5: Password Reset Service Logic

**Files:**
- Create: `internal/auth/password_reset.go`

- [ ] **Step 1: Create password_reset.go with token generation and service methods**

```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/hanasakis/kotoha/pkg/mail"
	"golang.org/x/crypto/bcrypt"
)

// generateResetToken creates a cryptographically random hex token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashResetToken returns the SHA256 hex digest of a token for storage.
func hashResetToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// ForgotPassword sends a password reset email if the email is registered.
// Always returns nil to prevent email enumeration.
func (s *Service) ForgotPassword(email string, mailCli *mail.Client) error {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		// User not found — return nil to prevent enumeration
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return err
	}

	now := time.Now()
	resetToken := &PasswordResetToken{
		TokenHash: hashResetToken(token),
		UserID:    user.ID,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	if err := s.repo.CreatePasswordResetToken(resetToken); err != nil {
		return err
	}

	return mailCli.SendPasswordReset(user.Email, token)
}

// ResetPassword validates the reset token and sets a new password.
func (s *Service) ResetPassword(token, newPassword string) error {
	tokenHash := hashResetToken(token)
	rt, err := s.repo.FindPasswordResetToken(tokenHash)
	if err != nil {
		return errors.New("auth.invalid_reset_token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user, err := s.repo.FindByID(rt.UserID)
	if err != nil {
		return errors.New("auth.invalid_reset_token")
	}
	user.PasswordHash = string(hash)
	if err := s.repo.DB.Save(user).Error; err != nil {
		return err
	}

	// Invalidate all existing sessions so old refresh tokens stop working
	s.repo.InvalidateUserSessions(user.ID)
	s.repo.MarkResetTokenUsed(rt.ID)

	return nil
}

// DeleteAccount verifies the password then soft-deletes the user.
func (s *Service) DeleteAccount(userID uint, password string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("auth.login_failed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("auth.invalid_password")
	}

	s.repo.InvalidateUserSessions(userID)
	return s.repo.DB.Delete(user).Error
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 6: Handlers & Routes

**Files:**
- Modify: `internal/auth/handler.go`
- Modify: `internal/router/router.go`

- [ ] **Step 1: Add ForgotPassword handler**

Append to `internal/auth/handler.go`:

```go
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	// mailCli is set on the handler by the router
	if err := h.svc.ForgotPassword(input.Email, h.mailCli); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "common.server_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
```

- [ ] **Step 2: Add ResetPassword handler**

```go
type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	if err := h.svc.ResetPassword(input.Token, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
```

- [ ] **Step 3: Add DeleteAccount handler**

```go
type DeleteAccountInput struct {
	Password string `json:"password" binding:"required"`
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var input DeleteAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "common.invalid_request"})
		return
	}
	if err := h.svc.DeleteAccount(userID, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
```

- [ ] **Step 4: Add `mailCli` field and `"github.com/hanasakis/kotoha/internal/middleware"` import to handler**

Update the `Handler` struct:

```go
type Handler struct {
	svc     *Service
	mailCli *mail.Client
}
```

Update `NewHandler`:

```go
func NewHandler(svc *Service, mailCli *mail.Client) *Handler {
	return &Handler{svc: svc, mailCli: mailCli}
}
```

Add imports at top:

```go
import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanasakis/kotoha/internal/middleware"
	"github.com/hanasakis/kotoha/pkg/mail"
)
```

- [ ] **Step 5: Register routes in router.go**

In the public auth group, add after `authGroup.POST("/refresh", authH.Refresh)`:

```go
authGroup.POST("/forgot-password", authH.ForgotPassword)
authGroup.POST("/reset-password", authH.ResetPassword)
```

In the protected group, add after `protected.POST("/auth/logout", authH.Logout)`:

```go
protected.DELETE("/auth/account", authH.DeleteAccount)
```

- [ ] **Step 6: Create mail client in router Setup() and pass to auth handler**

Add import `"github.com/hanasakis/kotoha/pkg/mail"` to router.go.

After `authRepo := auth.NewRepository(deps.DB)` and before `authSvc := ...`, add:

```go
mailCli := mail.New(
	deps.Config.SMTP.Host,
	deps.Config.SMTP.Port,
	deps.Config.SMTP.User,
	deps.Config.SMTP.Password,
	deps.Config.SMTP.From,
	deps.Config.SMTP.FromName,
)
```

Change `authH := auth.NewHandler(authSvc)` to:

```go
authH := auth.NewHandler(authSvc, mailCli)
```

- [ ] **Step 7: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 7: i18n Keys

**Files:**
- Modify: `internal/i18n/locales/zh.json`
- Modify: `internal/i18n/locales/en.json`

- [ ] **Step 1: Add zh keys**

Add before the closing `}` in `zh.json`:

```json
"auth.invalid_reset_token": "重置链接无效或已过期",
"auth.invalid_password": "当前密码错误",
"auth.account_deleted": "账号已注销"
```

- [ ] **Step 2: Add en keys**

Add before the closing `}` in `en.json`:

```json
"auth.invalid_reset_token": "Reset link is invalid or has expired",
"auth.invalid_password": "Current password is incorrect",
"auth.account_deleted": "Account has been deleted"
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 8: Auto-Migration

**Files:**
- Modify: `pkg/db/migration.go:17`

- [ ] **Step 1: Add PasswordResetToken to auto-migration list**

Change line 18 in `pkg/db/migration.go` — add after `&auth.Session{},`:

```go
&auth.User{},
&auth.Session{},
&auth.PasswordResetToken{},
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 9: Verify & Test

- [ ] **Step 1: Run `go vet`**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 2: Run existing tests**

Run: `go test ./internal/auth/... -v -count=1` (if auth tests exist)
Expected: existing tests pass (new features don't break)

- [ ] **Step 3: Verify all routes are registered**

Run: `go build ./...`
Expected: success, no unused imports
