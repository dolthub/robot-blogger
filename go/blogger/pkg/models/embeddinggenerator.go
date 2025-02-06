package models

import "context"

type EmbeddingGenerator interface {
	GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error)
}
