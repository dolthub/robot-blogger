package models

import "context"

type ModelServer interface {
	Chatter
	EmbeddingGenerator

	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
