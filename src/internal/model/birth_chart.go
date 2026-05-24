package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BirthChart stores a user's BaZi (八字) birth chart analysis.
type BirthChart struct {
	gorm.Model
	UserID       uint   `gorm:"not null;index" json:"user_id"`
	User         User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Name         string `gorm:"type:varchar(64);not null" json:"name"`
	Gender       string `gorm:"type:varchar(4);not null" json:"gender"` // 男/女
	BirthYear    int    `gorm:"not null" json:"birth_year"`
	BirthMonth   int    `gorm:"not null" json:"birth_month"`
	BirthDay     int    `gorm:"not null" json:"birth_day"`
	BirthHour    int    `gorm:"not null" json:"birth_hour"`
	BirthMin     int    `gorm:"not null;default:0" json:"birth_min"`
	CalendarType string `gorm:"type:varchar(16);not null;default:'solar'" json:"calendar_type"` // solar/lunar

	YearPillar  datatypes.JSON `gorm:"type:json" json:"year_pillar"`
	MonthPillar datatypes.JSON `gorm:"type:json" json:"month_pillar"`
	DayPillar   datatypes.JSON `gorm:"type:json" json:"day_pillar"`
	HourPillar  datatypes.JSON `gorm:"type:json" json:"hour_pillar"`

	FiveElements  datatypes.JSON `gorm:"type:json" json:"five_elements"`
	ElementDetail datatypes.JSON `gorm:"type:json" json:"element_detail"`
	BodyStrength  datatypes.JSON `gorm:"type:json" json:"body_strength"`
	TenGods       datatypes.JSON `gorm:"type:json" json:"ten_gods"`
	NaYin         datatypes.JSON `gorm:"type:json" json:"na_yin"`
	DaYunStart    datatypes.JSON `gorm:"type:json" json:"da_yun_start"`

	ZiWeiResult   datatypes.JSON `gorm:"type:json" json:"ziwei_result"`
	ZiWeiComputed bool           `gorm:"default:false" json:"ziwei_computed"`
}
