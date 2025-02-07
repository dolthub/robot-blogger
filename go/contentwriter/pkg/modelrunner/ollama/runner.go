package ollama

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/modelrunner"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

type ollamaAPIRunner struct {
	model  string
	cli    *api.Client
	mr     *api.ProcessModelResponse
	logger *zap.Logger
}

var _ modelrunner.ModelRunner = &ollamaAPIRunner{}

func NewOllamaLocallyRunningServer(model string, logger *zap.Logger) (*ollamaAPIRunner, error) {
	if os.Getenv("OLLAMA_HOST") == "" {
		return nil, fmt.Errorf("OLLAMA_HOST is not set")
	}

	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &ollamaAPIRunner{
		model:  model,
		cli:    cli,
		logger: logger,
	}, nil
}

func (s *ollamaAPIRunner) Start(ctx context.Context) error {
	// todo: make a request to the running ollama server,
	// error if response status is not 200
	running, err := s.cli.ListRunning(ctx)
	if err != nil {
		return err
	}

	if len(running.Models) == 0 {
		return fmt.Errorf("no running models found")
	}

	for _, r := range running.Models {
		parts := strings.Split(r.Name, ":")
		if parts[0] == s.model {
			s.mr = &r
			return nil
		}
	}

	return fmt.Errorf("model %s not found: make sure it is running", s.model)
}

func (s *ollamaAPIRunner) Stop(ctx context.Context) error {
	s.mr = nil
	return nil
}

func (s *ollamaAPIRunner) Chat(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("ollama api chat", zap.String("model", s.model), zap.String("prompt", prompt), zap.Duration("duration", time.Since(start)))
	}()

	if wc == nil {
		return 0, nil
	}

	stream := false

	req := &api.ChatRequest{
		Model:  s.model,
		Stream: &stream,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	var m int64
	respfn := func(resp2 api.ChatResponse) error {
		n, err := io.Copy(wc, strings.NewReader(resp2.Message.Content))
		if err != nil {
			return err
		}
		m += n
		return nil
	}

	err := s.cli.Chat(context.Background(), req, respfn)
	if err != nil {
		return 0, err
	}
	return m, nil
}

func (s *ollamaAPIRunner) ChatWithRAG(ctx context.Context, prompt, ragContent string, wc io.WriteCloser) (int64, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("ollama api chat with embeddings", zap.String("model", s.model), zap.String("prompt", prompt), zap.Duration("duration", time.Since(start)))
	}()

	if wc == nil {
		return 0, nil
	}

	stream := false

	req := &api.ChatRequest{
		Model:  s.model,
		Stream: &stream,
		Messages: []api.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(`Using the given reference text, answer the question that follows:
reference text is:

%s

end of reference text. The question is:

%s
				`, ragContent, prompt),
			},
		},
	}

	var m int64
	respfn := func(resp2 api.ChatResponse) error {
		n, err := io.Copy(wc, strings.NewReader(resp2.Message.Content))
		if err != nil {
			return err
		}
		m += n
		return nil
	}

	err := s.cli.Chat(context.Background(), req, respfn)
	if err != nil {
		return 0, err
	}
	return m, nil
}

func (s *ollamaAPIRunner) GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("ollama api generate embeddings", zap.String("model", s.model), zap.Duration("duration", time.Since(start)))
	}()

	req := &api.EmbeddingRequest{
		Model:  s.model,
		Prompt: prompt,
	}
	resp, err := s.cli.Embeddings(context.Background(), req)
	if err != nil {
		return nil, err
	}

	e := make([]float32, len(resp.Embedding))
	for i, f := range resp.Embedding {
		e[i] = float32(f)
	}

	return e, nil
}

func (s *ollamaAPIRunner) GetModelName() string {
	return s.model
}

func (s *ollamaAPIRunner) GetModelVersion() string {
	if s.mr == nil {
		return ""
	}
	return s.mr.Digest
}

func (s *ollamaAPIRunner) GetModelDimension() int {
	return 4096
}
