package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	lgdolt "github.com/tmc/langchaingo/vectorstores/dolt"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"go.uber.org/zap"
)

type DocSourceType string

const (
	DocSourceTypeBlogPost DocSourceType = "blog_post"
	DocSourceTypeEmail    DocSourceType = "email"
	DocSourceTypeCustom   DocSourceType = "custom"
)

type bloggerImpl struct {
	dst             DocSourceType
	llm             llms.Model
	s               vectorstores.VectorStore
	splitter        textsplitter.TextSplitter
	includeFileFunc func(path string) bool
	runner          Runner
	model           Model
	logger          *zap.Logger
}

var _ Blogger = &bloggerImpl{}

func NewBlogger(
	ctx context.Context,
	runner Runner,
	model Model,
	storeType StoreType,
	host string,
	port int,
	storeName string,
	splitter textsplitter.TextSplitter,
	includeFileFunc func(path string) bool,
	dst DocSourceType,
	logger *zap.Logger,
) (Blogger, error) {
	var err error
	var e *embeddings.EmbedderImpl

	var llm llms.Model
	switch runner {
	case OllamaRunner:
		llm, err = ollama.New(ollama.WithModel(string(model)))
		if err != nil {
			return nil, err
		}
		llmClient, ok := llm.(embeddings.EmbedderClient)
		if !ok {
			return nil, fmt.Errorf("llm does not implement embeddings.EmbedderClient")
		}
		e, err = embeddings.NewEmbedder(llmClient)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported runner: %s", runner)
	}
	if err != nil {
		return nil, err
	}

	var s vectorstores.VectorStore
	switch storeType {
	case Postgres:
		url := getPostgresURL("postgres", "", host, storeName, port)
		s, err = pgvector.New(
			ctx,
			pgvector.WithConnectionURL(url),
			pgvector.WithEmbedder(e),
		)
		if err != nil {
			return nil, err
		}
	case Dolt:
		url := getDoltURL("root", "", host, storeName, port)
		s, err = lgdolt.New(ctx,
			lgdolt.WithConnectionURL(url),
			lgdolt.WithEmbedder(e),
			lgdolt.WithCreateEmbeddingIndexAfterAddDocuments(true))
	default:
		return nil, fmt.Errorf("unsupported store: %s", storeType)
	}
	if err != nil {
		return nil, err
	}

	return &bloggerImpl{
		s:               s,
		llm:             llm,
		splitter:        splitter,
		includeFileFunc: includeFileFunc,
		dst:             dst,
		runner:          runner,
		model:           model,
		logger:          logger,
	}, nil
}

func (b *bloggerImpl) Store(ctx context.Context, dir string) error {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if b.includeFileFunc(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		contentHash, err := b.contentMd5(content)
		if err != nil {
			return err
		}

		md := map[string]any{
			"doc_source_type": string(b.dst),
			"name":            filepath.Base(file),
			"runner":          string(b.runner),
			"model":           string(b.model),
			"md5":             contentHash,
		}

		// TODO: check if content has already been added
		// TODO: if so, skip it

		docs, err := textsplitter.CreateDocuments(b.splitter, []string{string(content)}, []map[string]any{md})
		if err != nil {
			return err
		}

		start := time.Now()
		_, err = b.s.AddDocuments(ctx, docs)
		if err != nil {
			return err
		}

		b.logger.Info("finished storing document", zap.String("doc_source_type", string(b.dst)), zap.String("name", filepath.Base(file)), zap.Duration("duration", time.Since(start)))
	}

	return nil
}

func (b *bloggerImpl) Generate(ctx context.Context, prompt string, numSearchDocs int) error {
	docs, err := b.s.SimilaritySearch(ctx, prompt, numSearchDocs)
	if err != nil {
		return err
	}

	fullPrompt := prompt
	if len(docs) > 0 {
		fullPrompt = "Use the following pieces of context to answer the question at the end. The context pieces are as follows:\n"
		for idx, doc := range docs {
			fullPrompt += "context piece " + strconv.Itoa(idx+1) + ": \n"
			fullPrompt += fmt.Sprintf("%s\n", doc.PageContent)
			fullPrompt += "end of context piece " + strconv.Itoa(idx+1) + "\n\n"
		}
		fullPrompt += "The question is: " + prompt + "\n\n"
	}

	_, err = llms.GenerateFromSinglePrompt(
		ctx,
		b.llm,
		fullPrompt,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	return err
}

func (b *bloggerImpl) contentMd5(data []byte) (string, error) {
	r := bytes.NewReader(data)
	hash := md5.New()
	_, err := io.Copy(hash, r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func getPostgresURL(user, password, host, databaseName string, port int) string {
	if password == "" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s", user, host, port, databaseName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, databaseName)
}

func getDoltURL(user, password, host, databaseName string, port int) string {
	if password == "" {
		return fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, host, port, databaseName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, password, host, port, databaseName)
}
