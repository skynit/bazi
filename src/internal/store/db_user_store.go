package store

import (
	"bazi/internal/model"

	"gorm.io/gorm"
)

type DBUserStore struct {
	db *gorm.DB
}

func NewDBUserStore(db *gorm.DB) *DBUserStore {
	return &DBUserStore{db: db}
}

func (s *DBUserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *DBUserStore) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *DBUserStore) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *DBUserStore) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
