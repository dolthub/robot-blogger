package modelrunner

import (
	"context"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"io"
)

type noopModelRunner struct{}

func (n noopModelRunner) Chat(ctx context.Context, input string, wc io.WriteCloser) (int64, error) {
	return 0, nil
}

func (n noopModelRunner) ChatWithEmbeddings(ctx context.Context, input string, db dbs.DatabaseServer, wc io.WriteCloser) (int64, error) {
	return 0, nil
}

func (n noopModelRunner) GenerateEmbeddings(ctx context.Context, prompt string) ([]float32, error) {
	return []float32{}, nil
}

func (n noopModelRunner) Start(ctx context.Context) error {
	return nil
}

func (n noopModelRunner) Stop(ctx context.Context) error {
	return nil
}

func (n noopModelRunner) GetModelName() string {
	return ""
}

func (n noopModelRunner) GetModelVersion() string {
	return ""
}

func (n noopModelRunner) GetModelDimension() int {
	return 0
}

var _ ModelRunner = &noopModelRunner{}
