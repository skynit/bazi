package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FortuneQuery stores a daily fortune query linked to a birth chart.
type FortuneQuery struct {
	gorm.Model
	UserID  uint       `gorm:"not null;index" json:"user_id"`
	User    User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ChartID uint       `gorm:"not null;index" json:"chart_id"`
	Chart   BirthChart `gorm:"foreignKey:ChartID;constraint:OnDelete:CASCADE" json:"-"`

	QueryDate    time.Time      `gorm:"type:date;not null" json:"query_date"`
	DayPillar    datatypes.JSON `gorm:"type:json" json:"day_pillar"`
	DailyFortune datatypes.JSON `gorm:"type:json" json:"daily_fortune"`
}
