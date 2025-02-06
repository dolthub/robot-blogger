package llama3

import (
	"context"
	"io"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
)

type llama3WithEmbeddingsBlogger struct {
	ms models.ModelServer
	db dbs.DatabaseServer
}

var _ blogger.BlogWriterWithEmbeddings = &llama3WithEmbeddingsBlogger{}

func NewLlama3BloggerWithEmbeddings(ctx context.Context, ms models.ModelServer, db dbs.DatabaseServer) (*llama3WithEmbeddingsBlogger, error) {
	err := db.Start(ctx)
	if err != nil {
		return nil, err
	}

	err = ms.Start(ctx)
	if err != nil {
		return nil, err
	}

	modelName := ms.GetModelName()
	modelVersion := ms.GetModelVersion()
	modelDimension := ms.GetModelDimension()

	err = db.InsertModel(ctx, modelName, modelVersion, modelDimension)
	if err != nil {
		return nil, err
	}

	return &llama3WithEmbeddingsBlogger{
		ms: ms,
		db: db,
	}, nil
}

func (b *llama3WithEmbeddingsBlogger) UpdateInput(ctx context.Context, input blogger.Input) error {
	content := input.Content()
	embeddings, err := b.ms.GenerateEmbeddings(ctx, content)
	if err != nil {
		return err
	}
	contentMd5, err := input.ContentMd5()
	if err != nil {
		return err
	}
	return b.db.InsertEmbedding(ctx, input.ID(), b.ms.GetModelName(), b.ms.GetModelVersion(), contentMd5, content, embeddings)
}

func (b *llama3WithEmbeddingsBlogger) WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	return b.ms.ChatWithEmbeddings(ctx, prompt, b.db, wc)
}

func (b *llama3WithEmbeddingsBlogger) Close(ctx context.Context) error {
	err := b.db.Stop(ctx)
	if err != nil {
		return err
	}
	return b.ms.Stop(ctx)
}
