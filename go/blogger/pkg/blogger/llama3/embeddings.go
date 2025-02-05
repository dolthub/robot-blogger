package llama3

import (
	"context"
	"io"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
)

type llama3WithEmbeddingsBlogger struct {
	ms models.ModelServer
}

var _ blogger.BlogWriterWithEmbeddings = &llama3WithEmbeddingsBlogger{}

func NewLlama3BloggerWithEmbeddings(ms models.ModelServer) *llama3WithEmbeddingsBlogger {
	return &llama3WithEmbeddingsBlogger{
		ms: ms,
	}
}

func (b *llama3WithEmbeddingsBlogger) UpdateInput(ctx context.Context, input blogger.Input) error {
	return b.ms.GenerateEmbeddings(ctx, input.ID())
}

func (b *llama3WithEmbeddingsBlogger) WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	return b.ms.Chat(ctx, prompt, wc)
}
