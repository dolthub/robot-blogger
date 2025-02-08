package modelrunner

import (
	"context"
	"io"
)

type Chatter interface {
	Chat(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error)
	ChatWithRAG(ctx context.Context, prompt, ragContent string, wc io.WriteCloser) (int64, error)
}

type EmbeddingGenerator interface {
	GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error)
}

type ModelRunner interface {
	Chatter
	EmbeddingGenerator

	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetModelName() string
	GetModelVersion() string
	GetModelDimension() int
}
