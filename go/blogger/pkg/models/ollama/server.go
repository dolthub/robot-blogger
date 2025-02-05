package models

import (
	"context"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
	"github.com/ollama/ollama/api"
)

// this is a server that is locally running ollama
// ollama is expected to be running on the local machine
// and the model is expected to be locally available and running
type ollamaLocallyRunningServer struct {
	model string
	cli   *api.Client
}

var _ models.ModelServer = &ollamaLocallyRunningServer{}

func NewOllamaLocallyRunningServer(model string) (*ollamaLocallyRunningServer, error) {
	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &ollamaLocallyRunningServer{
		model: model,
		cli:   cli,
	}, nil
}

func (s *ollamaLocallyRunningServer) Start(ctx context.Context) error {
	// todo: make sure OLLAMA_HOST is set
	// todo: make a request to the locally running ollama server,
	// error if response status is not 200
	return nil
}

func (s *ollamaLocallyRunningServer) Stop(ctx context.Context) error {
	return nil
}

func (s *ollamaLocallyRunningServer) Chat(ctx context.Context, prompt string) (string, error) {
	return "", nil
}

func (s *ollamaLocallyRunningServer) GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error) {
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
