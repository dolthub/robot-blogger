package llama3

import (
	"context"
	"database/sql"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
	"io"
	"time"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/modelrunner"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer"
)

type llama3RagImpl struct {
	ms     modelrunner.ModelRunner
	db     dbs.DatabaseServer
	logger *zap.Logger
}

var _ writer.RAGContentReadWriter = &llama3RagImpl{}

func NewLlama3(ctx context.Context, ms modelrunner.ModelRunner, db dbs.DatabaseServer, logger *zap.Logger) (*llama3RagImpl, error) {
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

	if db.Name() == dbs.Postgres {
		err = createDBSchemaPostgres(ctx, db, logger)
		if err != nil {
			return nil, err
		}
	} else {
		panic("unsupported")
	}

	err = insertModelIfNotExists(ctx, db, modelName, modelVersion, modelDimension, logger)
	if err != nil {
		return nil, err
	}

	return &llama3RagImpl{
		ms:     ms,
		db:     db,
		logger: logger,
	}, nil
}

func createDBSchemaPostgres(ctx context.Context, db dbs.DatabaseServer, logger *zap.Logger) error {
	start := time.Now()
	defer func() {
		logger.Info("create db schema", zap.Duration("duration", time.Since(start)))
	}()
	err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS models (name text not null, version text not null, dimension int not null, created_at timestamp not null default current_timestamp, primary key (name, version))")
	if err != nil {
		return err
	}
	return db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS dolthub_blog_embeddings (id text, model_name_fk text not null, model_version_fk text not null, doc_index int not null, embedding vector(4096) not null, content_md5 text not null, content text not null, created_at timestamp not null default current_timestamp, primary key(id, content_md5, doc_index), foreign key (model_name_fk, model_version_fk) references models(name, version))")
}

func insertModelIfNotExists(ctx context.Context, db dbs.DatabaseServer, model string, version string, dimension int, logger *zap.Logger) error {
	start := time.Now()
	defer func() {
		logger.Info("write model metadata", zap.String("model", model), zap.String("version", version), zap.Int("dimension", dimension), zap.Duration("duration", time.Since(start)))
	}()

	if db.Name() == dbs.Postgres {
		return db.ExecContext(ctx, "INSERT INTO models (name, version, dimension) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", model, version, dimension)
	}
	panic("unsupported")
}

func (b *llama3RagImpl) UpdateInput(ctx context.Context, input writer.Input) error {
	var id string
	var model string
	var version string
	var docIndex int
	start := time.Now()
	defer func() {
		b.logger.Info("llama3 insert embedding", zap.String("id", id), zap.String("model", model), zap.String("version", version), zap.Int("doc_index", docIndex), zap.Duration("duration", time.Since(start)))
	}()

	content := input.Content()
	embedding, err := b.ms.GenerateEmbeddings(ctx, content)
	if err != nil {
		return err
	}
	contentMd5, err := input.ContentMd5()
	if err != nil {
		return err
	}

	if b.db.Name() == dbs.Postgres {
		return b.updateInputPostgres(ctx, id, model, version, contentMd5, content, docIndex, embedding)
	}
	panic("unsupported")
}

func (b *llama3RagImpl) updateInputPostgres(ctx context.Context, id, model, version, contentMd5, content string, docIndex int, embedding []float32) error {
	exists := false
	existsFunc := func(ctx context.Context, rows *sql.Rows) error {
		found := 0
		for rows.Next() {
			found++
		}
		if found > 0 {
			exists = true
		}
		return nil
	}
	if exists {
		return nil
	}

	// check if embedding already exists
	err := b.db.QueryContext(ctx, existsFunc, "SELECT * FROM dolthub_blog_embeddings WHERE id = $1 and content_md5 = $2 and doc_index = $3 and model_name_fk = $4 and model_version_fk = $5;", id, contentMd5, docIndex, model, version)
	if err != nil {
		return err
	}

	return b.db.ExecContext(ctx, "INSERT INTO dolthub_blog_embeddings (id, model_name_fk, model_version_fk, doc_index, embedding, content_md5, content) VALUES ($1, $2, $3, $4, $5, $6, $7)", id, model, version, docIndex, pgvector.NewVector(embedding), contentMd5, content)
}
func (b *llama3RagImpl) WriteContent(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	embeddings, err := b.ms.GenerateEmbeddings(ctx, prompt)
	if err != nil {
		return 0, err
	}

	ragContent, err := b.GetContentFromEmbeddings(ctx, embeddings)
	if err != nil {
		return 0, err
	}

	return b.ms.ChatWithRAG(ctx, prompt, ragContent, wc)
}

type Result struct {
	id      string
	content string
}

func (b *llama3RagImpl) getContentFromEmbeddingsFromPostgres(ctx context.Context, embeddings []float32) (string, error) {
	start := time.Now()
	defer func() {
		b.logger.Info("llama3 get content from embeddings", zap.Duration("duration", time.Since(start)))
	}()
	results := make([]Result, 0)

	getResultsFunc := func(ctx context.Context, rows *sql.Rows) error {
		for rows.Next() {
			var result Result
			err := rows.Scan(&result.id, &result.content)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		return nil
	}

	err := b.db.QueryContext(ctx, getResultsFunc, "SELECT id, content FROM dolthub_blog_embeddings ORDER BY embedding <-> $1 LIMIT 10", pgvector.NewVector(embeddings))
	if err != nil {
		return "", err
	}

	combinedContent := ""
	for _, result := range results {
		b.logger.Info("postgres locally running server get content from embeddings using id:", zap.String("id", result.id))
		combinedContent += result.content + "\n\n"
	}

	return combinedContent, nil
}

func (b *llama3RagImpl) GetContentFromEmbeddings(ctx context.Context, embeddings []float32) (string, error) {
	if b.db.Name() == dbs.Postgres {
		return b.getContentFromEmbeddingsFromPostgres(ctx, embeddings)
	}
	panic("unsupported")
}

func (b *llama3RagImpl) Close(ctx context.Context) error {
	err := b.db.Stop(ctx)
	if err != nil {
		return err
	}
	return b.ms.Stop(ctx)
}
