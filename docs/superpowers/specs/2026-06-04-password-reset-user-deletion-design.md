# Password Reset & User Deletion Design

> **Status:** Approved | **Date:** 2026-06-04

## Overview

Add password reset flow (via Resend email) and user account deletion to the Kotoha auth module.

---

## Feature 1: Password Reset

### Data Model

New table `password_reset_tokens`:

| Column | Type | Notes |
|--------|------|-------|
| id | uint PK | auto-increment |
| token_hash | string(64) unique | SHA256 of the random token |
| user_id | uint FK | references auth.users |
| expires_at | timestamp | 1 hour from creation |
| used | bool | default false |
| created_at | timestamp | |

### API

**`POST /api/v1/auth/forgot-password`**（public）

Request: `{"email": "user@example.com"}`

Logic:
1. Always return `{"message": "ok"}` — prevent email enumeration
2. Look up user by email; if not found, return early (no email sent)
3. Generate 32-byte random token via `crypto/rand`, encode as hex
4. Store `SHA256(token)` in `password_reset_tokens`
5. Send email with reset link: `https://kotoha.shop/reset-password?token=<token>`
6. Return `{"message": "ok"}`

**`POST /api/v1/auth/reset-password`**（public）

Request: `{"token": "...", "new_password": "..."}`

Logic:
1. Compute `SHA256(token)`, look up in DB where `used=false AND expires_at > now()`
2. If not found or expired → `{"code": "auth.invalid_token"}`
3. Hash new password with bcrypt, update user
4. Mark token as `used=true`
5. Invalidate all sessions for that user (force re-login)
6. Return `{"message": "ok"}`

### Email Delivery

New package `pkg/mail` — SMTP client using Resend:

```
Host: smtp.resend.com
Port: 587
Username: resend
Password: <SMTP_PASSWORD env var>
From: fushuinannan@gmail.com
FromName: Kotoha
```

**Config additions** (`internal/config/config.go`):

```go
type SMTPConfig struct {
    Host, Port, User, Password, From, FromName string
}
```

### Files

- **Create**: `pkg/mail/mail.go` — SMTP client with `Send(to, subject, html) error`
- **Create**: `internal/auth/password_reset.go` — token generation, email template, top-level orchestrator
- **Modify**: `internal/auth/models.go` — add `PasswordResetToken` model
- **Modify**: `internal/auth/repository.go` — add CRUD for reset tokens
- **Modify**: `internal/auth/service.go` — add `ForgotPassword`, `ResetPassword` methods
- **Modify**: `internal/auth/handler.go` — add handler methods
- **Modify**: `internal/config/config.go` — add `SMTP` config struct + env vars
- **Modify**: `internal/router/router.go` — register routes
- **Create**: `internal/auth/password_reset_test.go` — integration tests

### Error Codes (i18n)

| Key | zh | en |
|-----|----|----|
| `auth.invalid_reset_token` | 重置链接无效或已过期 | Reset link is invalid or has expired |
| `auth.reset_email_sent` | 如果该邮箱已注册，重置链接已发送 | If the email is registered, a reset link has been sent |

---

## Feature 2: User Deletion (Soft Delete)

### API

**`DELETE /api/v1/auth/account`**（protected, requires auth）

Request: `{"password": "current_password"}`

Logic:
1. Extract `user_id` from JWT via `middleware.UserIDFromContext(c)`
2. Look up user, verify bcrypt match on provided password
3. If password mismatch → `{"code": "auth.invalid_password"}`
4. Call `InvalidateUserSessions(userID)` — all sessions terminated
5. GORM soft-delete: `db.Delete(&user)` — sets `deleted_at`
6. Addresses, preferences, orders remain untouched (soft delete only affects users row)
7. Return `{"message": "ok"}`

### User Model

No changes needed — `auth.User` already has `DeletedAt gorm.DeletedAt`.

### Files

- **Modify**: `internal/auth/service.go` — add `DeleteAccount(userID uint, password string) error`
- **Modify**: `internal/auth/handler.go` — add `DeleteAccount` handler
- **Modify**: `internal/router/router.go` — add `protected.DELETE("/auth/account", ...)`
- **Modify**: `internal/auth/service_test.go` — add deletion test (if test file exists)

### Error Codes (i18n)

| Key | zh | en |
|-----|----|----|
| `auth.invalid_password` | 当前密码错误 | Current password is incorrect |
| `auth.account_deleted` | 账号已注销 | Account has been deleted |

---

## Route Registration

```go
// Public
v1.POST("/auth/forgot-password", authH.ForgotPassword)
v1.POST("/auth/reset-password", authH.ResetPassword)

// Protected — inside the existing protected group
protected.DELETE("/auth/account", authH.DeleteAccount)
```

---

## Test Plan

### Password Reset Tests

| Test | Scenario |
|------|----------|
| `TestForgotPassword_ExistingEmail` | Valid email → returns ok (email sent) |
| `TestForgotPassword_NonExistentEmail` | Invalid email → returns ok (no email sent, no leak) |
| `TestResetPassword_ValidToken` | Correct token → password updated, token marked used |
| `TestResetPassword_ExpiredToken` | Expired token → error |
| `TestResetPassword_AlreadyUsed` | Reused token → error |
| `TestResetPassword_InvalidToken` | Garbage token → error |
| `TestResetPassword_SessionsInvalidated` | After reset → old refresh tokens rejected |

### User Deletion Tests

| Test | Scenario |
|------|----------|
| `TestDeleteAccount_CorrectPassword` | Valid password → user soft-deleted, sessions gone |
| `TestDeleteAccount_WrongPassword` | Wrong password → error, user still exists |
| `TestDeleteAccount_CannotLoginAfter` | After deletion → login returns auth.login_failed |
| `TestDeleteAccount_OrdersPreserved` | After deletion → orders still queryable |
