package llama3

import (
	"context"
	"io"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/modelrunner"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer"
)

type llama3ContentWriter struct {
	ms modelrunner.ModelRunner
	db dbs.DatabaseServer
}

var _ writer.RAGContentWriter = &llama3ContentWriter{}

func NewLlama3ContentWriter(ctx context.Context, ms modelrunner.ModelRunner, db dbs.DatabaseServer) (*llama3ContentWriter, error) {
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

	return &llama3ContentWriter{
		ms: ms,
		db: db,
	}, nil
}

func (b *llama3ContentWriter) UpdateInput(ctx context.Context, input writer.Input) error {
	content := input.Content()
	embeddings, err := b.ms.GenerateEmbeddings(ctx, content)
	if err != nil {
		return err
	}
	contentMd5, err := input.ContentMd5()
	if err != nil {
		return err
	}
	return b.db.InsertEmbedding(ctx, input.ID(), b.ms.GetModelName(), b.ms.GetModelVersion(), contentMd5, content, embeddings, input.DocIndex())
}

func (b *llama3ContentWriter) WriteBlog(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	return b.ms.ChatWithEmbeddings(ctx, prompt, b.db, wc)
}

func (b *llama3ContentWriter) Close(ctx context.Context) error {
	err := b.db.Stop(ctx)
	if err != nil {
		return err
	}
	return b.ms.Stop(ctx)
}
