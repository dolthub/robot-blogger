package blogger

import (
	"context"
	"io"
)

type Blogger interface {
	UpdateInput(ctx context.Context, input Input) error
	WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error)
}
