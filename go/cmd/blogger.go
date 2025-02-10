package main

import "context"

type Runner string
type Model string
type Store string

const (
	OllamaRunner Runner = "ollama"
	OpenAIRunner Runner = "openai"
)

const (
	Llama3Model Model = "llama3"
)

const (
	PostgresStore Store = "postgres"
	DoltStore     Store = "dolt"
)

type Blogger interface {
	Store(ctx context.Context, dir string) error
	Generate(ctx context.Context, prompt string, numSearchDocs int) error
}
