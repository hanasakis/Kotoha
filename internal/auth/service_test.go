package auth_test

import (
	"strings"
	"testing"

	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
)

func TestRegisterAndLogin(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := auth.NewRepository(database)
	svc := auth.NewService(repo, "test-jwt-secret-for-testing-purposes", 15*60*1000000000, 720*60*60*1000000000)

	t.Run("register_new_user", func(t *testing.T) {
		result, err := svc.Register(auth.RegisterInput{
			Email:    "test@example.com",
			Password: "password123",
			Nickname: "tester",
		}, "ios", "127.0.0.1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.AccessToken == "" {
			t.Error("expected non-empty AccessToken")
		}
		if result.RefreshToken == "" {
			t.Error("expected non-empty RefreshToken")
		}
		if result.User.Email != "test@example.com" {
			t.Errorf("got email %q, want test@example.com", result.User.Email)
		}
		if result.User.Nickname != "tester" {
			t.Errorf("got nickname %q, want tester", result.User.Nickname)
		}
		if result.User.Role != "user" {
			t.Errorf("got role %q, want user", result.User.Role)
		}
		if result.User.PasswordHash == "" {
			t.Error("expected non-empty PasswordHash")
		}
		if result.User.PasswordHash == "password123" {
			t.Error("password should be hashed, not plaintext")
		}
	})

	t.Run("register_duplicate_email", func(t *testing.T) {
		_, err := svc.Register(auth.RegisterInput{
			Email:    "test@example.com",
			Password: "password456",
		}, "ios", "127.0.0.1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "email_exists") {
			t.Errorf("expected 'email_exists' error, got %q", err.Error())
		}
	})

	t.Run("login_correct_password", func(t *testing.T) {
		result, err := svc.Login(auth.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}, "android", "10.0.0.1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken == "" {
			t.Error("expected non-empty AccessToken")
		}
		if result.RefreshToken == "" {
			t.Error("expected non-empty RefreshToken")
		}
	})

	t.Run("login_wrong_password", func(t *testing.T) {
		_, err := svc.Login(auth.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}, "web", "10.0.0.1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "login_failed") {
			t.Errorf("expected 'login_failed' error, got %q", err.Error())
		}
	})

	t.Run("login_nonexistent_user", func(t *testing.T) {
		_, err := svc.Login(auth.LoginInput{
			Email:    "nobody@example.com",
			Password: "anything",
		}, "web", "10.0.0.1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "login_failed") {
			t.Errorf("expected 'login_failed' error, got %q", err.Error())
		}
	})
}

func TestRefreshAndLogout(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := auth.NewRepository(database)
	svc := auth.NewService(repo, "test-jwt-secret", 15*60*1000000000, 720*60*60*1000000000)

	result, err := svc.Register(auth.RegisterInput{
		Email:    "refresh@test.com",
		Password: "pass12345",
	}, "ios", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("refresh_valid_token", func(t *testing.T) {
		newResult, err := svc.Refresh(result.RefreshToken, "ios", "127.0.0.1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newResult.AccessToken == "" {
			t.Error("expected non-empty AccessToken")
		}
		if newResult.RefreshToken == result.RefreshToken {
			t.Error("expected new refresh token, got same one")
		}
	})

	t.Run("refresh_invalidated_token", func(t *testing.T) {
		_, err := svc.Refresh(result.RefreshToken, "ios", "127.0.0.1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("logout", func(t *testing.T) {
		loginRes, err := svc.Login(auth.LoginInput{
			Email:    "refresh@test.com",
			Password: "pass12345",
		}, "test", "127.0.0.1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err = svc.Logout(loginRes.RefreshToken)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
