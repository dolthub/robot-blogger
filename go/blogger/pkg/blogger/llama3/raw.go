package llama3

import (
	"context"
	"io"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
)

type llama3OnlyBlogger struct {
	ms models.ModelServer
}

var _ blogger.BlogWriter = &llama3OnlyBlogger{}

func NewLlama3OnlyBlogger(ms models.ModelServer) *llama3OnlyBlogger {
	return &llama3OnlyBlogger{
		ms: ms,
	}
}

func (b *llama3OnlyBlogger) WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	return b.ms.Chat(ctx, prompt, wc)
}
