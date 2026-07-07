package store

import (
	"bazi/internal/model"
	"time"

	"gorm.io/gorm"
)

type DBBuyiStore struct {
	db *gorm.DB
}

func NewDBBuyiStore(db *gorm.DB) *DBBuyiStore {
	return &DBBuyiStore{db: db}
}

func (s *DBBuyiStore) Create(record *model.BuyiRecord) error {
	return s.db.Create(record).Error
}

func (s *DBBuyiStore) FindByUserDate(userID uint, date time.Time) (*model.BuyiRecord, error) {
	var record model.BuyiRecord
	err := s.db.Where("user_id = ? AND divination_date = ?", userID, date).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}
