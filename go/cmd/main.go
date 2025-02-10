package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/zap"
)

var ollamaRunner = flag.Bool("ollama", true, "uses ollama llm runner")
var llama3Model = flag.Bool("llama3", true, "uses the llama3 model for generating the content")
var postgres = flag.Bool("postgres", false, "uses postgres to store embeddings")
var dolt = flag.Bool("dolt", false, "uses dolt to store embeddings")
var prompt = flag.String("prompt", "", "the prompt to run")
var storeBlogs = flag.Bool("store-blogs", false, "store dolthub blog documents")
var storeEmails = flag.Bool("store-emails", false, "store dolthub marketing email documents")
var storeCustom = flag.String("store-custom", "", "store custom documents")
var storeName = flag.String("store-name", "", "the name of the vector store to use")
var debug = flag.Bool("debug", false, "enable debug logging")

func main() {
	flag.Parse()

	var runner Runner
	var model Model
	var store Store

	if *ollamaRunner {
		runner = "ollama"
	} else {
		panic("unsupported runner")
	}

	if *llama3Model {
		model = "llama3"
	} else {
		panic("unsupported model")
	}

	if *postgres {
		store = "postgres"
	} else if *dolt {
		store = "dolt"
	} else {
		panic("unsupported store")
	}

	if *storeName == "" {
		panic("store name is required")
	}

	storeOnly := false
	var splitter textsplitter.TextSplitter
	var inputsDir string
	var includeFileFunc func(path string) bool

	if *storeBlogs {
		storeOnly = true

		// todo: make this configurable
		splitter = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(512),    // default is 512
			textsplitter.WithChunkOverlap(128), // default is 100
			textsplitter.WithCodeBlocks(true),
			textsplitter.WithHeadingHierarchy(true),
		)

		// todo: fix this to clone repo
		inputsDir = "/Users/dustin/src/ld/web/packages/blog/src/pages"

		includeFileFunc = func(path string) bool {
			return filepath.Ext(path) == ".md"
		}

	} else if *storeEmails {
		storeOnly = true
	} else if *storeCustom != "" {
		storeOnly = true
	}

	if splitter == nil {
		panic("unable to create textsplitter")
	}

	var err error
	ctx := context.Background()
	logger := zap.NewNop()
	if *debug {
		config := zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		logger, err = config.Build()
		if err != nil {
			panic(err)
		}
	}
	defer logger.Sync()

	start := time.Now()
	defer func() {
		logger.Info("blogger total time", zap.Duration("duration", time.Since(start)))
	}()

	blogger, err := NewBlogger(ctx, runner, model, store, *storeName, splitter, includeFileFunc, DocSourceTypeBlogPost, logger)
	if err != nil {
		logger.Error("error", zap.Error(err))
		os.Exit(1)
	}

	if storeOnly {
		err = blogger.Store(ctx, inputsDir)
	} else {
		err = blogger.Generate(ctx, *prompt, 10)
	}

	if err != nil {
		logger.Error("error", zap.Error(err))
		os.Exit(1)
	}
}
