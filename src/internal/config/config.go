package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all application configuration values.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	SQLitePath string
	UseSQLite  bool
	JWTSecret  string
	ServerPort string

	RAGEnabled        bool
	RAGProvider       string
	RAGTimeoutSeconds int
	RAGMinScore       float64
	RAGTopK           int

	LocalRAGIndexPath string
	LocalRAGSourceDir string

	RAGFlowEnabled        bool
	RAGFlowBaseURL        string
	RAGFlowAPIKey         string
	RAGFlowBaziDatasetID  string
	RAGFlowTimeoutSeconds int
	RAGFlowMinScore       float64
	RAGFlowTopK           int
}

// Load reads configuration from environment variables, applying defaults.
// Uses mock-mode defaults (no MySQL) when DB_HOST is not set.
func Load() *Config {
	dbHost := getEnv("DB_HOST", "")
	useSQLite := dbHost == ""
	sqlitePath := getEnv("SQLITE_PATH", "./data/bazi.db")
	ragProvider := getEnv("RAG_PROVIDER", "")
	if ragProvider == "" {
		ragProvider = "sqlite_fts5"
		if getEnvBool("RAGFLOW_ENABLED", false) {
			ragProvider = "ragflow"
		}
	}
	ragEnabled := getEnvBool("RAG_ENABLED", getEnvBool("RAGFLOW_ENABLED", false))
	ragTimeout := getEnvInt("RAG_TIMEOUT_SECONDS", getEnvInt("RAGFLOW_TIMEOUT_SECONDS", 8))
	ragMinScore := getEnvFloat("RAG_MIN_SCORE", getEnvFloat("RAGFLOW_MIN_SCORE", 0.35))
	ragTopK := getEnvInt("RAG_TOP_K", getEnvInt("RAGFLOW_TOP_K", 8))
	return &Config{
		DBHost:     dbHost,
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPass:     getEnv("DB_PASS", ""),
		DBName:     getEnv("DB_NAME", "bazi"),
		SQLitePath: sqlitePath,
		UseSQLite:  useSQLite,
		JWTSecret:  requireJWTSecret(),
		ServerPort: getEnv("SERVER_PORT", "8088"),

		RAGEnabled:        ragEnabled,
		RAGProvider:       ragProvider,
		RAGTimeoutSeconds: ragTimeout,
		RAGMinScore:       ragMinScore,
		RAGTopK:           ragTopK,

		LocalRAGIndexPath: getEnv("LOCAL_RAG_INDEX_PATH", "../data/bazi_fts.db"),
		LocalRAGSourceDir: getEnv("LOCAL_RAG_SOURCE_DIR", "/home/skynit/mingli_db/md/bazi"),

		RAGFlowEnabled:        getEnvBool("RAGFLOW_ENABLED", ragEnabled && ragProvider == "ragflow"),
		RAGFlowBaseURL:        getEnv("RAGFLOW_BASE_URL", ""),
		RAGFlowAPIKey:         getEnv("RAGFLOW_API_KEY", ""),
		RAGFlowBaziDatasetID:  getEnv("RAGFLOW_BAZI_DATASET_ID", ""),
		RAGFlowTimeoutSeconds: ragTimeout,
		RAGFlowMinScore:       ragMinScore,
		RAGFlowTopK:           ragTopK,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func requireJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[WARN] JWT_SECRET is not set. Using auto-generated key. Set JWT_SECRET env var for production.")
		secret = "auto-generated-" + os.Getenv("HOSTNAME")
		if secret == "auto-generated-" {
			secret = "auto-generated-bazi-dev"
		}
	}
	return secret
}
