package modelrunner

import (
	"context"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"io"
)

type Chatter interface {
	Chat(ctx context.Context, input string, wc io.WriteCloser) (int64, error)
	ChatWithEmbeddings(ctx context.Context, input string, db dbs.DatabaseServer, wc io.WriteCloser) (int64, error)
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
