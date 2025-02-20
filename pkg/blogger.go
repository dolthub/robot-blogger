package pkg

import "context"

type Runner string
type Model string
type StoreType string

const (
	OllamaRunner Runner = "ollama"
	OpenAIRunner Runner = "openai"
)

const (
	Postgres StoreType = "postgres"
	MariaDB  StoreType = "mariadb"
	Dolt     StoreType = "dolt"
)

type Blogger interface {
	Store(ctx context.Context, dir string) error
	Generate(ctx context.Context, userPrompt string, topic string, length int, outputFormat string) error
	Close() error
}
