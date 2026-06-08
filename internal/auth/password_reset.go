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
// Returns the reset token (for dev mode display) and an error.
// Always returns nil error to prevent email enumeration.
func (s *Service) ForgotPassword(email string, mailCli *mail.Client) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		// User not found — return empty to prevent enumeration
		return "", "", nil
	}

	token, err := generateResetToken()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	resetToken := &PasswordResetToken{
		TokenHash: hashResetToken(token),
		UserID:    user.ID,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	if err := s.repo.CreatePasswordResetToken(resetToken); err != nil {
		return "", "", err
	}

	resetURL := "http://localhost:3000/reset-password?token=" + token
	if err := mailCli.SendPasswordReset(user.Email, token); err != nil {
		// Email failed (dev mode / no SMTP) — return the reset URL in the response
		return token, resetURL, nil
	}
	// Email sent — don't expose the reset URL
	return "", "", nil
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
