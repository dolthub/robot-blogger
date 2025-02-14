package pkg

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
	"strings"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	lgdolt "github.com/tmc/langchaingo/vectorstores/dolt"
	lgmd "github.com/tmc/langchaingo/vectorstores/mariadb"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"go.uber.org/zap"
)

type DocSourceType string

const (
	DocSourceTypeBlogPost      DocSourceType = "blog_post"
	DocSourceTypeEmail         DocSourceType = "email"
	DocSourceTypeDocumentation DocSourceType = "documentation"
	DocSourceTypeCustom        DocSourceType = "custom"
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

	var s vectorstores.VectorStore
	switch config.StoreType {
	case Postgres:
		url := GetPostgresConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		s, err = pgvector.New(
			ctx,
			pgvector.WithConnectionURL(url),
			pgvector.WithEmbedder(e),
		)
		if err != nil {
			return nil, err
		}
	case Dolt:
		url := GetDoltConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		s, err = lgdolt.New(ctx,
			lgdolt.WithConnectionURL(url),
			lgdolt.WithEmbedder(e),
			lgdolt.WithCreateEmbeddingIndexAfterAddDocuments(true))
	case MariaDB:
		url := GetMariaDBConnectionString(config.User, config.Password, config.Host, config.StoreName, config.Port)
		s, err = lgmd.New(ctx,
			lgmd.WithConnectionURL(url),
			lgmd.WithEmbedder(e),
			lgmd.WithVectorDimensions(config.VectorDimensions))
	default:
		return nil, fmt.Errorf("unsupported vector store: %s", config.StoreType)
	}
	if err != nil {
		return nil, err
	}

	return &bloggerImpl{
		s:               s,
		llm:             llm,
		splitter:        config.Splitter,
		includeFileFunc: config.IncludeFileFunc,
		dst:             config.DocSourceType,
		runner:          config.Runner,
		model:           config.Model,
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

		// todo: put the blog tags in the metadata
		// todo: put other useful shit in the metadata
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

var SystemPromptPreContentBlock = `
You are an expert content writer specializing in technical writing and marketing writing about Dolt, DoltHub and its related products.
Use the provided context document(s) to write new content based on the user's prompt.
You should write in a style that is engaging and informative, and to the point.
You should not copy the context verbatim, but rather use it as a guide to write new, engaging content.
Be sure to introduce new perspectives and ideas. Also, try to match the company's style and voice.
Each context document will be indicated by the following start and end tags:

<context>
</context>

The user prompt will be indicated by the following start and end tags:

<user_prompt>
</user_prompt>

The topic of your content will be indicated by the following start and end tags:

<topic>
</topic>

The length of your content will be indicated by the following start and end tags:

<length>
</length>

The output format of your content will be indicated by the following start and end tags:

<output_format>
</output_format>

Here are the context documents:

`

var SystemPromptPostContentBlock = `
Here are the topic, length, user's prompt, and output format:

<topic>
%s
</topic>

<length>
%d
</length>

<user_prompt>
%s
</user_prompt>

<output_format>
%s
</output_format>
`

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

	// todo: add a second retriever and merge the results
	// Use retriever to fetch relevant documents
	retrieverResult, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			b.llm,
			vectorstores.ToRetriever(
				b.s,
				numSearchDocs,
			),
		),
		userPrompt,
	)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(SystemPromptPreContentBlock)
	sb.WriteString(fmt.Sprintf("<context>%s</context>\n", retrieverResult))
	sb.WriteString(fmt.Sprintf(SystemPromptPostContentBlock, topic, length, userPrompt, outputFormat))
	systemPrompt := sb.String()

	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: systemPrompt}},
	}

	// Generate the final content
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
