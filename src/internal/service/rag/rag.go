package rag

import (
	"context"
	"errors"
)

var (
	ErrDisabled        = errors.New("rag disabled")
	ErrNotConfigured   = errors.New("rag not configured")
	ErrTimeout         = errors.New("rag timeout")
	ErrUpstream        = errors.New("rag upstream error")
	ErrInvalidResponse = errors.New("rag invalid response")
)

type RetrieveRequest struct {
	Question string
	Focus    string
}

type RetrievedChunk struct {
	ID         string
	Content    string
	Score      float64
	DocumentID string
	Metadata   map[string]string
}

type Retriever interface {
	Retrieve(ctx context.Context, req RetrieveRequest) ([]RetrievedChunk, error)
}
