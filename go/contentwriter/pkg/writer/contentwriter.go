package writer

import (
	"context"
	"io"
)

type Closer interface {
	Close(ctx context.Context) error
}

type InputUpdater interface {
	UpdateInput(ctx context.Context, input Input) error
	Closer
}

type ContentWriter interface {
	WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error)
	Closer
}

type RAGContentWriter interface {
	InputUpdater
	ContentWriter
}
