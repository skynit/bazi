package model

import (
	"time"

	"gorm.io/gorm"
)

// FortuneRecord stores a user's fortune calculation history.
type FortuneRecord struct {
	gorm.Model
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	ChartID     uint      `gorm:"not null;index" json:"chart_id"`
	QueryDate   time.Time `gorm:"not null" json:"query_date"`
	PeriodType  string    `gorm:"type:varchar(16);not null" json:"period_type"` // daily, weekly, monthly
	Pillar      string    `gorm:"type:json" json:"pillar"`                     // JSON
	Score       int       `gorm:"default:0" json:"score"`
	FiveElement string    `gorm:"type:json" json:"five_element"`               // JSON
	Luck        string    `gorm:"type:json" json:"luck"`                       // JSON
	Summary     string    `gorm:"type:text" json:"summary"`
}