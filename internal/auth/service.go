package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
	accessTTL time.Duration
	refreshTTL time.Duration
}

func NewService(repo *Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResult struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s *Service) Register(input RegisterInput, deviceType, deviceIP string) (*AuthResult, error) {
	// Check including soft-deleted users — if a deleted account exists,
	// permanently remove it so the email can be reused.
	existing, err := s.repo.FindByEmail(input.Email)
	if err == nil && existing != nil {
		return nil, errors.New("auth.email_exists")
	}
	// Also check soft-deleted: if found, permanently delete to free the email
	if deleted, err := s.repo.FindByEmailUnscoped(input.Email); err == nil && deleted != nil {
		s.repo.HardDeleteUser(deleted.ID)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        input.Email,
		PasswordHash: string(hash),
		Nickname:     input.Nickname,
		Role:         "user",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return s.generateAuthResult(user, deviceType, deviceIP)
}

func (s *Service) Login(input LoginInput, deviceType, deviceIP string) (*AuthResult, error) {
	user, err := s.repo.FindByEmail(input.Email)
	if err != nil {
		return nil, errors.New("auth.login_failed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("auth.login_failed")
	}

	return s.generateAuthResult(user, deviceType, deviceIP)
}

func (s *Service) Refresh(refreshToken, deviceType, deviceIP string) (*AuthResult, error) {
	session, err := s.repo.FindSessionByRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("auth.invalid_token")
	}

	if time.Now().After(session.ExpiresAt) {
		s.repo.InvalidateSession(session.ID)
		return nil, errors.New("auth.invalid_token")
	}

	user, err := s.repo.FindByID(session.UserID)
	if err != nil {
		return nil, errors.New("auth.invalid_token")
	}

	s.repo.InvalidateSession(session.ID)

	return s.generateAuthResult(user, deviceType, deviceIP)
}

func (s *Service) Logout(refreshToken string) error {
	session, err := s.repo.FindSessionByRefreshToken(refreshToken)
	if err != nil {
		return nil
	}
	return s.repo.InvalidateSession(session.ID)
}

func (s *Service) generateAuthResult(user *User, deviceType, deviceIP string) (*AuthResult, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTTL)

	claims := jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", user.ID),
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  expiresAt.Unix(),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)

	dt := deviceType
	if len(dt) > 500 {
		dt = dt[:500]
	}
	session := &Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		DeviceType:   dt,
		DeviceIP:     deviceIP,
		IsValid:      true,
		ExpiresAt:    now.Add(s.refreshTTL),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}
