package dbs

import "context"

type noopDatabaseServer struct {
}

var _ DatabaseServer = &noopDatabaseServer{}

func NewNoopDatabaseServer() *noopDatabaseServer {
	return &noopDatabaseServer{}
}

func (s *noopDatabaseServer) InsertModel(ctx context.Context, model string, version string, dimension int) error {
	return nil
}

func (s *noopDatabaseServer) InsertEmbedding(ctx context.Context, id, model, version, contentMd5, content string, embedding []float32, docIndex int) error {
	return nil
}

func (s *noopDatabaseServer) Start(ctx context.Context) error {
	return nil
}

func (s *noopDatabaseServer) Stop(ctx context.Context) error {
	return nil
}

func (s *noopDatabaseServer) GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error) {
	return "", nil
}
