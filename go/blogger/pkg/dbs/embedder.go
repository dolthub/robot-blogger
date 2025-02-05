package dbs

import (
	"context"
)

type Embedder interface {
	Embed(ctx context.Context, input []float32) error
}
