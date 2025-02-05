package dbs

import "context"

type noopDatabaseServer struct {
}

var _ DatabaseServer = &noopDatabaseServer{}

func NewNoopDatabaseServer() *noopDatabaseServer {
	return &noopDatabaseServer{}
}

func (s *noopDatabaseServer) Embed(ctx context.Context, input []float32) error {
	return nil
}

func (s *noopDatabaseServer) Start(ctx context.Context) error {
	return nil
}

func (s *noopDatabaseServer) Stop(ctx context.Context) error {
	return nil
}
