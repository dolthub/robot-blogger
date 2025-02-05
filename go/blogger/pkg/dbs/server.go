package dbs

import (
	"context"
)

type DatabaseServer interface {
	Embedder

	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
