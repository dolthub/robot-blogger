package main

import "context"

type Runner string
type Model string
type StoreType string

const (
	OllamaRunner Runner = "ollama"
	OpenAIRunner Runner = "openai"
)

const (
	Llama3Model Model = "llama3"
)

const (
	Postgres StoreType = "postgres"
	Dolt     StoreType = "dolt"
)

type Blogger interface {
	Store(ctx context.Context, dir string) error
	Generate(ctx context.Context, prompt string, numSearchDocs int) error
}
