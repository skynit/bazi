package model

import (
	"time"

	"gorm.io/gorm"
)

// BuyiRecord stores one daily divination result per user.
type BuyiRecord struct {
	gorm.Model
	UserID         uint      `gorm:"not null;uniqueIndex:idx_buyi_user_date" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	DivinationDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_buyi_user_date" json:"divination_date"`
	HexagramNumber int       `gorm:"not null" json:"hexagram_number"`
	HexagramName   string    `gorm:"type:varchar(32);not null" json:"hexagram_name"`
	Score          int       `gorm:"not null" json:"score"`
	Level          string    `gorm:"type:varchar(16);not null" json:"level"`
	Summary        string    `gorm:"type:text;not null" json:"summary"`
	HumanWay       string    `gorm:"type:text;not null" json:"human_way"`
	ImageReading   string    `gorm:"type:text;not null" json:"image_reading"`
	Advice         string    `gorm:"type:text;not null" json:"advice"`
}

type BuyiRecordResponse struct {
	ID             uint   `json:"id"`
	HexagramNumber int    `json:"hexagram_number"`
	HexagramName   string `json:"hexagram_name"`
	Summary        string `json:"summary"`
	HumanWay       string `json:"human_way"`
	ImageReading   string `json:"image_reading"`
	Advice         string `json:"advice"`
	Source         string `json:"source"`
	CreatedAt      string `json:"created_at"`
}

type BuyiTodayResponse struct {
	Date         string              `json:"date"`
	HasRecord    bool                `json:"has_record"`
	AlreadyDrawn bool                `json:"already_drawn"`
	Record       *BuyiRecordResponse `json:"record"`
}
