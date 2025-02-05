package blogger

import (
	"context"
	"io"
)

type InputUpdater interface {
	UpdateInput(ctx context.Context, input Input) error
}

type BlogWriter interface {
	WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error)
}

type BlogWriterWithEmbeddings interface {
	InputUpdater
	BlogWriter
}
