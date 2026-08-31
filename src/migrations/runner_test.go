package migrations

import (
	"strings"
	"testing"

	"bazi/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "-")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	return db
}

func TestApplyCreatesCurrentSchemaOnce(t *testing.T) {
	db := openTestDatabase(t)

	if err := Apply(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	if err := Apply(db); err != nil {
		t.Fatalf("reapply schema: %v", err)
	}

	for _, table := range []any{
		&model.User{},
		&model.BirthChart{},
		&model.FortuneRecord{},
		&model.FortuneFeedback{},
		&model.BuyiRecord{},
		&model.ElementAsset{},
		&model.AuspiciousRule{},
		&model.ActivityCatalog{},
	} {
		if !db.Migrator().HasTable(table) {
			t.Errorf("missing migrated table for %T", table)
		}
	}

	var versions []schemaMigration
	if err := db.Order("version").Find(&versions).Error; err != nil {
		t.Fatalf("list schema versions: %v", err)
	}
	if len(versions) != 1 || versions[0].Version != CurrentVersion {
		t.Fatalf("schema versions = %+v, want version %d exactly once", versions, CurrentVersion)
	}
}

func TestApplyRejectsNewerSchema(t *testing.T) {
	db := openTestDatabase(t)
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		t.Fatalf("create schema migration table: %v", err)
	}
	if err := db.Create(&schemaMigration{Version: CurrentVersion + 1, Name: "future"}).Error; err != nil {
		t.Fatalf("seed future schema version: %v", err)
	}

	err := Apply(db)
	if err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("Apply error = %v, want newer-schema rejection", err)
	}
}
