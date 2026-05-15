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

// Load reads configuration from environment variables, applying defaults where appropriate.
// Panics if JWT_SECRET is not set.
func Load() *Config {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", "123456"),
		DBName:     getEnv("DB_NAME", "bazi"),
		JWTSecret:  getEnv("JWT_SECRET", "1234567890abcdef"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
