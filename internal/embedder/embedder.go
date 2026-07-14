package embedder

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Embedder interface {
	Embed(ctx context.Context, text string) (Embedding, error)
}

type OpenAIEmbedder struct {
	client openai.Client
}

type Embedding struct {
	Values []float32
}

func NewEmbedder(apiKey string, baseURL string) *OpenAIEmbedder {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	client := openai.NewClient(opts...)
	return &OpenAIEmbedder{
		client: client,
	}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, content string) (Embedding, error) {
	embedding, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{content},
		},
		Model: openai.EmbeddingModelTextEmbedding3Small,
	})
	if err != nil {
		return Embedding{}, err
	}
	
	// Convert []float64 to []float32
	float64Vals := embedding.Data[0].Embedding
	float32Vals := make([]float32, len(float64Vals))
	for i, v := range float64Vals {
		float32Vals[i] = float32(v)
	}

	return Embedding{
		Values: float32Vals,
	}, nil
}
