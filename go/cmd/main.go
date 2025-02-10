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
var numDocs = flag.Int("num-docs", 100, "number of RAG documents to retrieve during content generation")

func main() {
	flag.Parse()

	var runner Runner
	var model Model
	var sn StoreType

	if *ollamaRunner {
		runner = OllamaRunner
	} else {
		panic("unsupported runner")
	}

	if *llama3Model {
		model = Llama3Model
	} else {
		panic("unsupported model")
	}

	if *postgres {
		sn = Postgres
	} else if *dolt {
		sn = Dolt
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
	} else {
		splitter = NewNoopTextSplitter()
	}

	if !storeOnly {
		if *numDocs == 0 {
			panic("number of documents must be greater than zero")
		}
	}

	var err error
	logger := zap.NewNop()
	if storeOnly {
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	}

	ctx := context.Background()

	start := time.Now()
	defer func() {
		logger.Info("blogger total time", zap.Duration("duration", time.Since(start)))
	}()

	blogger, err := NewBlogger(ctx, runner, model, sn, *storeName, splitter, includeFileFunc, DocSourceTypeBlogPost, logger)
	if err != nil {
		logger.Error("error", zap.Error(err))
		os.Exit(1)
	}

	if storeOnly {
		err = blogger.Store(ctx, inputsDir)
	} else {
		err = blogger.Generate(ctx, *prompt, *numDocs)
	}
	if err != nil {
		logger.Error("error", zap.Error(err))
		os.Exit(1)
	}
}
