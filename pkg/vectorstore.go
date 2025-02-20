package pkg

import (
	"context"

	"github.com/tmc/langchaingo/vectorstores"
)

type HasableVectorStore interface {
	Has(ctx context.Context, metadata map[string]any) (bool, error)
	Close() error
	vectorstores.VectorStore
}
