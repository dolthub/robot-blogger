package models

import (
	"context"
)

// this is a server that is locally running ollama
// ollama is expected to be running on the local machine
// and the model is expected to be locally available and running
type ollamaLocallyRunningServer struct {
	model string
	port  int
	host  string
}

func (s *ollamaLocallyRunningServer) Start(ctx context.Context) error {
	return nil
}

func (s *ollamaLocallyRunningServer) Stop(ctx context.Context) error {
	return nil
}
