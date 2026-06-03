package auth

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByID(id uint) (*User, error) {
	var user User
	err := r.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(user *User) error {
	return r.DB.Create(user).Error
}

func (r *Repository) CreateSession(session *Session) error {
	return r.DB.Create(session).Error
}

func (r *Repository) FindSessionByRefreshToken(token string) (*Session, error) {
	var session Session
	err := r.DB.Where("refresh_token = ? AND is_valid = ?", token, true).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) InvalidateSession(sessionID uint) error {
	return r.DB.Model(&Session{}).Where("id = ?", sessionID).Update("is_valid", false).Error
}

func (r *Repository) InvalidateUserSessions(userID uint) error {
	return r.DB.Model(&Session{}).Where("user_id = ? AND is_valid = ?", userID, true).Update("is_valid", false).Error
}

func (r *Repository) CleanExpiredSessions() error {
	return r.DB.Where("expires_at < ?", time.Now()).Delete(&Session{}).Error
}
