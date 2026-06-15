package ragflow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bazi/internal/service/rag"
)

func TestParseRetrievalResponse(t *testing.T) {
	body := []byte(`{
		"data": {
			"chunks": [
				{
					"id": "c1",
					"content": "  第一段  ",
					"similarity": 0.88,
					"metadata": {"book": "子平真诠", "chapter": "001", "source_path": "bazi/子平真诠/001.md"}
				}
			]
		}
	}`)
	chunks, err := parseRetrievalResponse(body)
	if err != nil {
		t.Fatalf("parseRetrievalResponse error: %v", err)
	}
	if len(chunks) != 1 || chunks[0].Content != "第一段" || chunks[0].Metadata["book"] != "子平真诠" {
		t.Fatalf("unexpected chunks: %+v", chunks)
	}
}

func TestClientRetrieve(t *testing.T) {
	var gotAuth string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"chunks":[{"id":"1","content":"hello","similarity":0.9,"metadata":{"book":"A","chapter":"1","source_path":"a.md"}}]}}`))
	}))
	defer srv.Close()

	client := NewClient(Config{
		Enabled:        true,
		BaseURL:        srv.URL,
		APIKey:         "abc",
		DatasetID:      "ds1",
		TimeoutSeconds: 1,
		MinScore:       0.5,
		TopK:           2,
	})
	chunks, err := client.Retrieve(context.Background(), rag.RetrieveRequest{Question: "q", Focus: "overview"})
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	if gotAuth != "Bearer abc" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if !contains(gotBody, `"dataset_ids":["ds1"]`) || !contains(gotBody, `"question":"q"`) || !contains(gotBody, `"top_k":2`) || !contains(gotBody, `"logic":"and"`) || !contains(gotBody, `"comparison_operator":"="`) {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
	if len(chunks) != 1 || chunks[0].Content != "hello" {
		t.Fatalf("unexpected chunks: %+v", chunks)
	}
}

func TestClientDisabled(t *testing.T) {
	client := NewClient(Config{})
	_, err := client.Retrieve(context.Background(), rag.RetrieveRequest{})
	if err == nil || err != ErrDisabled {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}

func TestClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(Config{
		Enabled:        true,
		BaseURL:        srv.URL,
		APIKey:         "abc",
		DatasetID:      "ds1",
		TimeoutSeconds: 0,
	})
	client.httpClient.Timeout = 1 * time.Nanosecond
	_, err := client.Retrieve(context.Background(), rag.RetrieveRequest{})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestParseRetrievalResponseUpstreamCode(t *testing.T) {
	_, err := parseRetrievalResponse([]byte(`{"code":100,"message":"dataset not found","data":{"chunks":[]}}`))
	if err == nil || !errors.Is(err, ErrUpstream) {
		t.Fatalf("expected upstream error, got %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > len(sub) && (stringIndex(s, sub) >= 0))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
