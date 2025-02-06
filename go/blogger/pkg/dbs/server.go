package dbs

import (
	"context"
)

type DatabaseServer interface {
	InsertModel(ctx context.Context, model string, version string, dimension int) error
	InsertEmbedding(ctx context.Context, id, model, version, content string, embedding []float32) error
	GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
