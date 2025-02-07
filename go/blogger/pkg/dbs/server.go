package dbs

import (
	"context"
)

type DatabaseServer interface {
	InsertModel(ctx context.Context, model string, version string, dimension int) error
	InsertEmbedding(ctx context.Context, id, model, version, contentMd5, content string, embedding []float32, docIndex int) error
	GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
