package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASS")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("SERVER_PORT")

	cfg := Load()

	if cfg.DBHost != "" {
		t.Errorf("expected DBHost empty, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "3306" {
		t.Errorf("expected DBPort=3306, got %s", cfg.DBPort)
	}
	if cfg.DBUser != "root" {
		t.Errorf("expected DBUser=root, got %s", cfg.DBUser)
	}
	if cfg.DBPass != "" {
		t.Errorf("expected DBPass empty, got %s", cfg.DBPass)
	}
	if cfg.DBName != "bazi" {
		t.Errorf("expected DBName=bazi, got %s", cfg.DBName)
	}
	if cfg.ServerPort != "8088" {
		t.Errorf("expected ServerPort=8088, got %s", cfg.ServerPort)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWTSecret=test-secret, got %s", cfg.JWTSecret)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("JWT_SECRET", "override-secret")
	os.Setenv("DB_HOST", "prod-db")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASS", "s3cret")
	os.Setenv("DB_NAME", "bazi_prod")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASS")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := Load()

	if cfg.DBHost != "prod-db" {
		t.Errorf("expected DBHost=prod-db, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("expected DBPort=5432, got %s", cfg.DBPort)
	}
	if cfg.DBUser != "admin" {
		t.Errorf("expected DBUser=admin, got %s", cfg.DBUser)
	}
	if cfg.DBPass != "s3cret" {
		t.Errorf("expected DBPass=s3cret, got %s", cfg.DBPass)
	}
	if cfg.DBName != "bazi_prod" {
		t.Errorf("expected DBName=bazi_prod, got %s", cfg.DBName)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort=9090, got %s", cfg.ServerPort)
	}
	if cfg.JWTSecret != "override-secret" {
		t.Errorf("expected JWTSecret=override-secret, got %s", cfg.JWTSecret)
	}
}

func TestLoadJWTSecretDefault(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("JWT_SECRET")

	cfg := Load()
	if cfg.JWTSecret != "dev-secret" {
		t.Errorf("expected JWTSecret=dev-secret, got %s", cfg.JWTSecret)
	}
}
