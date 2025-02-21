package pkg

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
	lgdolt "github.com/tmc/langchaingo/vectorstores/dolt"
	lgmd "github.com/tmc/langchaingo/vectorstores/mariadb"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"go.uber.org/zap"
)

type DocSourceType string

type bloggerImpl struct {
	llm                     llms.Model
	s                       HasableVectorStore
	splitter                textsplitter.TextSplitter
	includeFileFunc         func(path string) bool
	runner                  Runner
	model                   Model
	logger                  *zap.Logger
	preContentSystemPrompt  string
	postContentSystemPrompt string
}

var _ Blogger = &bloggerImpl{}

func NewBlogger(
	ctx context.Context,
	config *Config,
	logger *zap.Logger,
) (Blogger, error) {
	var err error
	var e *embeddings.EmbedderImpl

	var llm llms.Model
	switch config.Runner {
	case OllamaRunner:
		llm, err = ollama.New(ollama.WithModel(string(config.Model)))
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
	case OpenAIRunner:
		llm, err = openai.New(openai.WithModel(string(config.Model)))
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
		return nil, fmt.Errorf("unsupported llm runner: %s", config.Runner)
	}
	if err != nil {
		return nil, err
	}

	var s HasableVectorStore
	switch config.StoreType {
	case Postgres:
		url := GetPostgresConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		vs, err := pgvector.New(
			ctx,
			pgvector.WithConnectionURL(url),
			pgvector.WithEmbedder(e),
		)
		if err != nil {
			return nil, err
		}

		s, err = NewPostgresHasableVectorStore(vs, url)
		if err != nil {
			return nil, err
		}
	case Dolt:
		url := GetDoltConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		vs, err := lgdolt.New(ctx,
			lgdolt.WithConnectionURL(url),
			lgdolt.WithEmbedder(e),
			lgdolt.WithCreateEmbeddingIndexAfterAddDocuments(true))
		if err != nil {
			return nil, err
		}

		s, err = NewDoltHasableVectorStore(vs, url)
		if err != nil {
			return nil, err
		}
	case MariaDB:
		url := GetMariaDBConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		vs, err := lgmd.New(ctx,
			lgmd.WithConnectionURL(url),
			lgmd.WithEmbedder(e),
			lgmd.WithVectorDimensions(config.VectorDimensions))
		if err != nil {
			return nil, err
		}

		s, err = NewMariaDBHasableVectorStore(vs, url)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported vector store: %s", config.StoreType)
	}

	return &bloggerImpl{
		s:                       s,
		llm:                     llm,
		splitter:                config.Splitter,
		includeFileFunc:         config.IncludeFileFunc,
		runner:                  config.Runner,
		model:                   config.Model,
		logger:                  logger,
		preContentSystemPrompt:  config.PreContentSystemPrompt,
		postContentSystemPrompt: config.PostContentSystemPrompt,
	}, nil
}

func (b *bloggerImpl) Store(ctx context.Context, docSourceType DocSourceType, dir string) error {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if b.includeFileFunc(path) {
			b.logger.Info("preparing to store file", zap.String("file", filepath.Base(path)))
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
			"doc_source_type": string(docSourceType),
			"name":            filepath.Base(file),
			"runner":          string(b.runner),
			"model":           string(b.model),
			"md5":             contentHash,
		}

		has, err := b.s.Has(ctx, md)
		if err != nil {
			return err
		}
		if has {
			b.logger.Info("document already exists", zap.String("doc_source_type", string(docSourceType)), zap.String("name", filepath.Base(file)))
			continue
		}

		docs, err := textsplitter.CreateDocuments(b.splitter, []string{string(content)}, []map[string]any{md})
		if err != nil {
			return err
		}

		start := time.Now()
		_, err = b.s.AddDocuments(ctx, docs)
		if err != nil {
			return err
		}

		b.logger.Info("finished storing document", zap.String("doc_source_type", string(docSourceType)), zap.String("name", filepath.Base(file)), zap.Duration("duration", time.Since(start)))
	}

	return nil
}

func (b *bloggerImpl) getNumSearchDocs(length int) int {
	if length < 300 {
		return 3
	} else if length < 2500 {
		return 4
	} else if length < 5000 {
		return 5
	} else if length < 7500 {
		return 6
	} else if length < 10000 {
		return 7
	}
	return 9
}

func (b *bloggerImpl) Generate(ctx context.Context, userPrompt string, topic string, length int, outputFormat string) error {
	numSearchDocs := b.getNumSearchDocs(length)

	docs, err := b.s.SimilaritySearch(ctx, userPrompt, numSearchDocs)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return errors.New("no relevant documents found")
	}
	var sb strings.Builder
	sb.WriteString(b.preContentSystemPrompt)
	for _, doc := range docs {
		fmt.Println()
		fmt.Println("DOC SIM SCORE: ", doc.Score)
		fmt.Println()
		sb.WriteString(fmt.Sprintf("\n<context>\n%s\n</context>\n", doc.PageContent))
	}
	sb.WriteString(fmt.Sprintf(b.postContentSystemPrompt, topic, length, userPrompt, outputFormat))
	systemPrompt := sb.String()

	fmt.Println()
	fmt.Println("FINAL SYSTEM PROMPT:")
	fmt.Println(systemPrompt)
	fmt.Println()

	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: systemPrompt}},
	}

	_, err = b.llm.GenerateContent(ctx,
		[]llms.MessageContent{msg},
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

func (b *bloggerImpl) Close() error {
	return b.s.Close()
}

func GetPostgresConnectionString(user, password, host, databaseName string, port int) string {
	if password == "" {
		return fmt.Sprintf("postgres://%s@%s:%d/%s", user, host, port, databaseName)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, databaseName)
}

func GetDoltConnectionString(user, password, host, databaseName string, port int) string {
	if password == "" {
		return fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, host, port, databaseName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, password, host, port, databaseName)
}

func GetMariaDBConnectionString(user, password, host, databaseName string, port int) string {
	if password == "" {
		return fmt.Sprintf("tcp(%s:%d)/%s", host, port, databaseName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, databaseName)
}
