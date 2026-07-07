package store

import (
	"testing"
	"time"

	"bazi/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBuyiStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.BuyiRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestDBBuyiStoreFindCreateAndUniqueDate(t *testing.T) {
	db := setupBuyiStoreTestDB(t)
	store := NewDBBuyiStore(db)
	date := time.Date(2026, 7, 5, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))

	found, err := store.FindByUserDate(1, date)
	if err != nil {
		t.Fatalf("find empty: %v", err)
	}
	if found != nil {
		t.Fatalf("expected nil record, got %+v", found)
	}

	record := &model.BuyiRecord{
		UserID:         1,
		DivinationDate: date,
		HexagramNumber: 11,
		HexagramName:   "地天泰",
		Score:          86,
		Level:          "大吉",
		Summary:        "今日得地天泰",
		HumanWay:       "上下交通，泰平盛世。",
		ImageReading:   "三人同行、喜报、日月同辉（大吉之象）。",
		Advice:         "今日气势较顺。",
	}
	if err := store.Create(record); err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err = store.FindByUserDate(1, date)
	if err != nil {
		t.Fatalf("find created: %v", err)
	}
	if found == nil || found.HexagramName != "地天泰" {
		t.Fatalf("expected created record, got %+v", found)
	}

	duplicate := *record
	duplicate.ID = 0
	if err := store.Create(&duplicate); err == nil {
		t.Fatal("expected duplicate user/date to fail")
	}

	var count int64
	if err := db.Model(&model.BuyiRecord{}).Where("user_id = ? AND divination_date = ?", 1, date).Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one record after duplicate create, got %d", count)
	}
}
