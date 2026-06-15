package ragflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bazi/internal/service/rag"
)

var (
	ErrDisabled        = rag.ErrDisabled
	ErrNotConfigured   = rag.ErrNotConfigured
	ErrTimeout         = rag.ErrTimeout
	ErrUpstream        = rag.ErrUpstream
	ErrInvalidResponse = rag.ErrInvalidResponse
)

type Config struct {
	Enabled        bool
	BaseURL        string
	APIKey         string
	DatasetID      string
	TimeoutSeconds int
	MinScore       float64
	TopK           int
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Retrieve(ctx context.Context, req rag.RetrieveRequest) ([]rag.RetrievedChunk, error) {
	if c == nil || !c.cfg.Enabled {
		return nil, ErrDisabled
	}
	if strings.TrimSpace(c.cfg.BaseURL) == "" || strings.TrimSpace(c.cfg.APIKey) == "" || strings.TrimSpace(c.cfg.DatasetID) == "" {
		return nil, ErrNotConfigured
	}

	topK := c.cfg.TopK
	if topK <= 0 {
		topK = 8
	}
	minScore := c.cfg.MinScore
	if minScore <= 0 {
		minScore = 0.35
	}

	payload := map[string]interface{}{
		"dataset_ids":          []string{c.cfg.DatasetID},
		"question":             req.Question,
		"top_k":                topK,
		"similarity_threshold": minScore,
		"metadata_condition": map[string]interface{}{
			"logic": "and",
			"conditions": []map[string]interface{}{
				{"name": "domain", "comparison_operator": "=", "value": "bazi"},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/v1/retrieval"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrUpstream, resp.StatusCode, truncateForError(string(respBody)))
	}

	chunks, err := parseRetrievalResponse(respBody)
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

type retrievalResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    retrievalData       `json:"data"`
	Chunks  []retrievalChunkRaw `json:"chunks"`
}

type retrievalData struct {
	Chunks []retrievalChunkRaw `json:"chunks"`
}

type retrievalChunkRaw struct {
	ID              string                 `json:"id"`
	ChunkID         string                 `json:"chunk_id"`
	Content         string                 `json:"content"`
	Text            string                 `json:"text"`
	DocumentID      string                 `json:"document_id"`
	DocumentKeyword string                 `json:"document_keyword"`
	Similarity      float64                `json:"similarity"`
	Score           float64                `json:"score"`
	Metadata        map[string]interface{} `json:"metadata"`
	MetaFields      map[string]interface{} `json:"meta_fields"`
}

func parseRetrievalResponse(body []byte) ([]rag.RetrievedChunk, error) {
	var raw retrievalResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("%w: code=%d message=%s", ErrUpstream, raw.Code, raw.Message)
	}

	rawChunks := raw.Data.Chunks
	if len(rawChunks) == 0 {
		rawChunks = raw.Chunks
	}

	chunks := make([]rag.RetrievedChunk, 0, len(rawChunks))
	for _, rc := range rawChunks {
		content := strings.TrimSpace(rc.Content)
		if content == "" {
			content = strings.TrimSpace(rc.Text)
		}
		id := rc.ID
		if id == "" {
			id = rc.ChunkID
		}
		score := rc.Similarity
		if score == 0 {
			score = rc.Score
		}
		metadata := normalizeMetadata(rc.Metadata)
		for k, v := range normalizeMetadata(rc.MetaFields) {
			if metadata[k] == "" {
				metadata[k] = v
			}
		}
		if metadata["book"] == "" && rc.DocumentKeyword != "" {
			metadata["book"] = rc.DocumentKeyword
		}

		chunks = append(chunks, rag.RetrievedChunk{
			ID:         id,
			Content:    content,
			Score:      score,
			DocumentID: rc.DocumentID,
			Metadata:   metadata,
		})
	}
	return chunks, nil
}

func normalizeMetadata(raw map[string]interface{}) map[string]string {
	out := map[string]string{}
	for k, v := range raw {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(fmt.Sprint(v))
	}
	return out
}

func truncateForError(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 200 {
		return s
	}
	return s[:200]
}
