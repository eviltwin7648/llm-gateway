package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const OPENAI_URL = "http/example/openai/embedding/url"

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type Embedder struct {
	client *http.Client
}
type Embedding struct {
	Id     string
	Values []float32
}

func NewEmbedder() *Embedder {
	return &Embedder{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *Embedder) Embed(ctx context.Context, content string) *Embedding {
	embedding := &Embedding{}
	client := e.client
	body, err := json.Marshal(content)
	resp, err := client.Post(OPENAI_URL, "application/json", bytes.NewReader(body))
	if err != nil {

	}
	defer resp.Body.Close()
	return embedding
}
