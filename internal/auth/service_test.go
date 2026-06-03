package auth_test

import (
	"testing"

	"github.com/hanasakis/kotoha/internal/auth"
	"github.com/hanasakis/kotoha/internal/testutil"
	"github.com/hanasakis/kotoha/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndLogin(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := auth.NewRepository(database)
	svc := auth.NewService(repo, "test-jwt-secret-for-testing-purposes", 3600, 7200)

	t.Run("register_new_user", func(t *testing.T) {
		result, err := svc.Register(auth.RegisterInput{
			Email:    "test@example.com",
			Password: "password123",
			Nickname: "tester",
		}, "ios", "127.0.0.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "test@example.com", result.User.Email)
		assert.Equal(t, "tester", result.User.Nickname)
		assert.Equal(t, "user", result.User.Role)
		assert.NotEmpty(t, result.User.PasswordHash)
		assert.NotEqual(t, "password123", result.User.PasswordHash)
	})

	t.Run("register_duplicate_email", func(t *testing.T) {
		_, err := svc.Register(auth.RegisterInput{
			Email:    "test@example.com",
			Password: "password456",
		}, "ios", "127.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email_exists")
	})

	t.Run("login_correct_password", func(t *testing.T) {
		result, err := svc.Login(auth.LoginInput{
			Email:    "test@example.com",
			Password: "password123",
		}, "android", "10.0.0.1")
		require.NoError(t, err)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
	})

	t.Run("login_wrong_password", func(t *testing.T) {
		_, err := svc.Login(auth.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}, "web", "10.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login_failed")
	})

	t.Run("login_nonexistent_user", func(t *testing.T) {
		_, err := svc.Login(auth.LoginInput{
			Email:    "nobody@example.com",
			Password: "anything",
		}, "web", "10.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login_failed")
	})
}

func TestRefreshAndLogout(t *testing.T) {
	database := testutil.SetupTestDB(t)
	defer testutil.CleanTestDB(t, database)

	if err := db.AutoMigrate(database); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := auth.NewRepository(database)
	svc := auth.NewService(repo, "test-jwt-secret", 3600, 7200)

	result, err := svc.Register(auth.RegisterInput{
		Email:    "refresh@test.com",
		Password: "pass12345",
	}, "ios", "127.0.0.1")
	require.NoError(t, err)

	t.Run("refresh_valid_token", func(t *testing.T) {
		newResult, err := svc.Refresh(result.RefreshToken, "ios", "127.0.0.1")
		require.NoError(t, err)
		assert.NotEmpty(t, newResult.AccessToken)
		assert.NotEqual(t, result.AccessToken, newResult.AccessToken)
		assert.NotEqual(t, result.RefreshToken, newResult.RefreshToken)
	})

	t.Run("refresh_invalidated_token", func(t *testing.T) {
		_, err := svc.Refresh(result.RefreshToken, "ios", "127.0.0.1")
		assert.Error(t, err)
	})

	t.Run("logout", func(t *testing.T) {
		loginRes, _ := svc.Login(auth.LoginInput{
			Email:    "refresh@test.com",
			Password: "pass12345",
		}, "test", "127.0.0.1")
		err := svc.Logout(loginRes.RefreshToken)
		assert.NoError(t, err)
	})
}
