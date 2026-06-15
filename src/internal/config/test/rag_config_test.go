package config_test

import (
	. "bazi/internal/config"
	"os"
	"testing"
)

func clearRAGEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"RAG_ENABLED",
		"RAG_PROVIDER",
		"RAG_TIMEOUT_SECONDS",
		"RAG_MIN_SCORE",
		"RAG_TOP_K",
		"LOCAL_RAG_INDEX_PATH",
		"LOCAL_RAG_SOURCE_DIR",
		"RAGFLOW_ENABLED",
		"RAGFLOW_BASE_URL",
		"RAGFLOW_API_KEY",
		"RAGFLOW_BAZI_DATASET_ID",
		"RAGFLOW_TIMEOUT_SECONDS",
		"RAGFLOW_MIN_SCORE",
		"RAGFLOW_TOP_K",
	} {
		os.Unsetenv(key)
	}
}

func TestLoadRAGDefaults(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")
	clearRAGEnv(t)

	cfg := Load()

	if cfg.RAGEnabled {
		t.Fatal("expected RAGEnabled=false by default")
	}
	if cfg.RAGProvider != "sqlite_fts5" {
		t.Fatalf("expected sqlite_fts5 provider, got %q", cfg.RAGProvider)
	}
	if cfg.RAGTimeoutSeconds != 8 {
		t.Fatalf("expected timeout 8, got %d", cfg.RAGTimeoutSeconds)
	}
	if cfg.RAGMinScore != 0.35 {
		t.Fatalf("expected min score 0.35, got %v", cfg.RAGMinScore)
	}
	if cfg.RAGTopK != 8 {
		t.Fatalf("expected topK 8, got %d", cfg.RAGTopK)
	}
	if cfg.LocalRAGIndexPath != "../data/bazi_fts.db" {
		t.Fatalf("unexpected local index path: %s", cfg.LocalRAGIndexPath)
	}
}

func TestLoadRAGOverrides(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("RAG_ENABLED", "true")
	os.Setenv("RAG_PROVIDER", "sqlite")
	os.Setenv("RAG_TIMEOUT_SECONDS", "12")
	os.Setenv("RAG_MIN_SCORE", "0.6")
	os.Setenv("RAG_TOP_K", "3")
	os.Setenv("LOCAL_RAG_INDEX_PATH", "/tmp/bazi_fts.db")
	os.Setenv("LOCAL_RAG_SOURCE_DIR", "/tmp/md")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		clearRAGEnv(t)
	}()

	cfg := Load()

	if !cfg.RAGEnabled || cfg.RAGProvider != "sqlite" || cfg.RAGTimeoutSeconds != 12 || cfg.RAGMinScore != 0.6 || cfg.RAGTopK != 3 || cfg.LocalRAGIndexPath != "/tmp/bazi_fts.db" || cfg.LocalRAGSourceDir != "/tmp/md" {
		t.Fatalf("unexpected rag config: %+v", cfg)
	}
}

func TestLoadRAGFlowCompatibility(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("RAGFLOW_ENABLED", "true")
	os.Setenv("RAGFLOW_BASE_URL", "https://rag.example")
	os.Setenv("RAGFLOW_API_KEY", "key")
	os.Setenv("RAGFLOW_BAZI_DATASET_ID", "dataset")
	os.Setenv("RAGFLOW_TIMEOUT_SECONDS", "11")
	os.Setenv("RAGFLOW_MIN_SCORE", "0.55")
	os.Setenv("RAGFLOW_TOP_K", "4")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		clearRAGEnv(t)
	}()

	cfg := Load()

	if !cfg.RAGEnabled || cfg.RAGProvider != "ragflow" {
		t.Fatalf("expected legacy ragflow env to enable ragflow provider: %+v", cfg)
	}
	if !cfg.RAGFlowEnabled || cfg.RAGFlowBaseURL != "https://rag.example" || cfg.RAGFlowAPIKey != "key" || cfg.RAGFlowBaziDatasetID != "dataset" {
		t.Fatalf("unexpected ragflow config: %+v", cfg)
	}
	if cfg.RAGTimeoutSeconds != 11 || cfg.RAGMinScore != 0.55 || cfg.RAGTopK != 4 {
		t.Fatalf("expected legacy score/time config to map to generic config: %+v", cfg)
	}
}
