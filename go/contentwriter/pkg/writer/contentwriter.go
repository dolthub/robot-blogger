package writer

import (
	"context"
	"io"
)

type Closer interface {
	Close(ctx context.Context) error
}

type ContentReader interface {
	GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error)
}

type InputUpdater interface {
	UpdateInput(ctx context.Context, input Input) error
	Closer
}

type ContentWriter interface {
	WriteContent(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error)
	Closer
}

type RAGContentReadWriter interface {
	InputUpdater
	ContentWriter
	ContentReader
}
