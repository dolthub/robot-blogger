package models

import (
	"context"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
	"github.com/ollama/ollama/api"
)

type ollamaAPI struct {
	model string
	cli   *api.Client
}

var _ models.ModelAPI = &ollamaAPI{}

func NewOllamaAPIFromEnvironment(ctx context.Context, model string) (*ollamaAPI, error) {
	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &ollamaAPI{cli: cli, model: model}, nil
}

func (s *ollamaAPI) Chat(ctx context.Context, prompt string) (string, error) {
	return "", nil
}

func (s *ollamaAPI) GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error) {
	doc := ""

	req := &api.EmbeddingRequest{
		Model:  s.model,
		Prompt: doc,
	}
	resp, err := s.cli.Embeddings(context.Background(), req)
	if err != nil {
		return nil, err
	}

	e := make([]float32, len(resp.Embedding))
	for i, f := range resp.Embedding {
		e[i] = float32(f)
	}

	return e, nil
}
