package models

import "context"

type EmbeddingGenerator interface {
	GenerateEmbeddings(ctx context.Context, input string) error
}
