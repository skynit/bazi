package config

import "os"

// Config holds all application configuration values.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	JWTSecret  string
	ServerPort string
}

// Load reads configuration from environment variables, applying defaults.
// Uses mock-mode defaults (no MySQL) when DB_HOST is not set.
func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", ""),
		DBName:     getEnv("DB_NAME", "bazi"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret"),
		ServerPort: getEnv("SERVER_PORT", "8088"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
