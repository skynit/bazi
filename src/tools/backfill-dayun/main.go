package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"bazi/internal/config"
	"bazi/internal/model"
	"bazi/internal/service/bazi"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	apply := flag.Bool("apply", false, "write updates; default is dry-run")
	sqlitePath := flag.String("sqlite-path", "", "SQLite database path; overrides SQLITE_PATH")
	flag.Parse()

	cfg := config.Load()
	if *sqlitePath != "" {
		cfg.SQLitePath = *sqlitePath
	}
	if cfg.UseSQLite {
		cfg.SQLitePath = resolveSQLitePath(cfg.SQLitePath)
	}
	db := openDB(cfg)
	svc := &bazi.BaziService{}

	var charts []model.BirthChart
	if err := db.Find(&charts).Error; err != nil {
		log.Fatalf("query charts: %v", err)
	}

	updated := 0
	skipped := 0
	candidates := 0
	for i := range charts {
		chart := &charts[i]
		if hasJSONValue(chart.DaYunStart) {
			continue
		}
		candidates++
		result, err := svc.Calculate(chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.BirthMin, chart.Gender)
		if err != nil {
			log.Printf("skip chart %d: %v", chart.ID, err)
			skipped++
			continue
		}
		data, err := json.Marshal(result.DaYunInfo)
		if err != nil {
			log.Printf("skip chart %d: marshal dayun: %v", chart.ID, err)
			skipped++
			continue
		}
		if *apply {
			var current model.BirthChart
			if err := db.Select("id", "da_yun_start").First(&current, chart.ID).Error; err != nil {
				log.Printf("skip chart %d: reload: %v", chart.ID, err)
				skipped++
				continue
			}
			if hasJSONValue(current.DaYunStart) {
				skipped++
				continue
			}
			if err := db.Model(chart).Update("da_yun_start", datatypes.JSON(data)).Error; err != nil {
				log.Printf("skip chart %d: update: %v", chart.ID, err)
				skipped++
				continue
			}
		}
		updated++
	}

	mode := "dry-run"
	if *apply {
		mode = "apply"
	}
	fmt.Printf("backfill-dayun mode=%s charts=%d candidates=%d updated=%d skipped=%d\n", mode, len(charts), candidates, updated, skipped)
}

func hasJSONValue(data datatypes.JSON) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null"))
}

func resolveSQLitePath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if path == "./data/bazi.db" || path == "data/bazi.db" {
		if _, err := os.Stat("../data/bazi.db"); err == nil {
			return "../data/bazi.db"
		}
	}
	return path
}

func openDB(cfg *config.Config) *gorm.DB {
	var dialector gorm.Dialector
	if cfg.UseSQLite {
		dialector = sqlite.Open(cfg.SQLitePath)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	return db
}
