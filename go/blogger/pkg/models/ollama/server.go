package models

import (
	"context"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
)

// this is a server that is locally running ollama
// ollama is expected to be running on the local machine
// and the model is expected to be locally available and running
type ollamaLocallyRunningServer struct {
	model string
	port  int
	host  string
}

var _ models.ModelServer = &ollamaLocallyRunningServer{}

func NewOllamaLocallyRunningServer(model string, port int) *ollamaLocallyRunningServer {
	return &ollamaLocallyRunningServer{
		model: model,
		port:  port,
		host:  "127.0.0.1",
	}
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
