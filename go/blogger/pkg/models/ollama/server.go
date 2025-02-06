package ollama

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
	"github.com/ollama/ollama/api"
)

// this is a server that is locally running ollama
// ollama is expected to be running on the local machine
// and the model is expected to be locally available and running
type ollamaLocallyRunningServer struct {
	model string
	cli   *api.Client
	mr    *api.ProcessModelResponse
}

var _ models.ModelServer = &ollamaLocallyRunningServer{}

func NewOllamaLocallyRunningServer(model string) (*ollamaLocallyRunningServer, error) {
	if os.Getenv("OLLAMA_HOST") == "" {
		return nil, fmt.Errorf("OLLAMA_HOST is not set")
	}

	cli, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &ollamaLocallyRunningServer{
		model: model,
		cli:   cli,
	}, nil
}

func (s *ollamaLocallyRunningServer) Start(ctx context.Context) error {
	// todo: make a request to the locally running ollama server,
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

func (s *ollamaLocallyRunningServer) Stop(ctx context.Context) error {
	s.mr = nil
	return nil
}

func (s *ollamaLocallyRunningServer) Chat(ctx context.Context, prompt string, wc io.WriteCloser) (int64, error) {
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

func (s *ollamaLocallyRunningServer) ChatWithEmbeddings(ctx context.Context, prompt string, db dbs.DatabaseServer, wc io.WriteCloser) (int64, error) {
	if wc == nil {
		return 0, nil
	}

	embeddings, err := s.GenerateEmbeddings(ctx, prompt)
	if err != nil {
		return 0, err
	}

	content, err := db.GetContentFromEmbeddings(ctx, embeddings)
	if err != nil {
		return 0, err
	}

	fmt.Printf("** using content: %s\n", content[:30])

	stream := false

	req := &api.ChatRequest{
		Model:  s.model,
		Stream: &stream,
		Messages: []api.Message{
			{
				Role: "user",
				Content: fmt.Sprintf(`Using the given reference text, succinctly answer the question that follows:
reference text is:

%s

end of reference text. The question is:

%s
				`, content, prompt),
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

	err = s.cli.Chat(context.Background(), req, respfn)
	if err != nil {
		return 0, err
	}
	return m, nil
}

func (s *ollamaLocallyRunningServer) GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error) {
	doc := ""

	req := &api.EmbeddingRequest{
		Model:  s.model,
		Prompt: doc,
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

func (s *ollamaLocallyRunningServer) GetModelName() string {
	return s.model
}

func (s *ollamaLocallyRunningServer) GetModelVersion() string {
	if s.mr == nil {
		return ""
	}
	return s.mr.Digest
}

func (s *ollamaLocallyRunningServer) GetModelDimension() int {
	return 4096
}
