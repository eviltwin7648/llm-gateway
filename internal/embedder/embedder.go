package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

const OPENAI_URL = "http/example/openai/embedding/url"

type Embedder interface {
	Embed(ctx context.Context, text string) (Embedding, error)
}

type OpenAIEmbedder struct {
	client *http.Client
}
type Embedding struct {
	Id     string
	Values []float32
}
type embeddingResponse struct { //ref openaidocs
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}
}

func NewEmbedder() *OpenAIEmbedder {
	return &OpenAIEmbedder{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, content string) (Embedding, error) {
	client := e.client
	body, err := json.Marshal(content)
	resp, err := client.Post(OPENAI_URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return Embedding{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Embedding{}, err
	}
	var result embeddingResponse
	err = json.Unmarshal(raw, &result)
	if err != nil {
		return Embedding{}, err
	}
	if result.Error != nil {
		return Embedding{}, errors.New(result.Error.Message)
	}

	return Embedding{
		Id:     "",
		Values: result.Data[0].Embedding,
	}, nil
}
