package models

import (
	"context"
	"io"
)

type Chatter interface {
	Chat(ctx context.Context, input string, wc io.WriteCloser) (int64, error)
}
