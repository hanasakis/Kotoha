package user

import (
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

func (s *Service) GetProfile(userID uint) (*Profile, error) {
	var profile Profile
	err := s.DB.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return &Profile{UserID: userID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *Service) UpsertProfile(profile *Profile) error {
	var existing Profile
	err := s.DB.Where("user_id = ?", profile.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.DB.Create(profile).Error
	}
	if err != nil {
		return err
	}
	profile.ID = existing.ID
	return s.DB.Save(profile).Error
}

func (s *Service) GetAddresses(userID uint) ([]Address, error) {
	var addresses []Address
	err := s.DB.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addresses).Error
	return addresses, err
}

func (s *Service) CreateAddress(addr *Address) error {
	if addr.IsDefault {
		s.DB.Model(&Address{}).Where("user_id = ?", addr.UserID).Update("is_default", false)
	}
	return s.DB.Create(addr).Error
}

func (s *Service) UpdateAddress(addr *Address) error {
	if addr.IsDefault {
		s.DB.Model(&Address{}).Where("user_id = ? AND id != ?", addr.UserID, addr.ID).Update("is_default", false)
	}
	return s.DB.Save(addr).Error
}

func (s *Service) DeleteAddress(userID, addressID uint) error {
	return s.DB.Where("id = ? AND user_id = ?", addressID, userID).Delete(&Address{}).Error
}

func (s *Service) GetPreference(userID uint) (*Preference, error) {
	var pref Preference
	err := s.DB.Where("user_id = ?", userID).First(&pref).Error
	if err == gorm.ErrRecordNotFound {
		return &Preference{UserID: userID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (s *Service) UpsertPreference(pref *Preference) error {
	var existing Preference
	err := s.DB.Where("user_id = ?", pref.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.DB.Create(pref).Error
	}
	if err != nil {
		return err
	}
	pref.ID = existing.ID
	return s.DB.Save(pref).Error
}
