package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dolthub/robot-blogger/pkg"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/zap"
)

var help = flag.Bool("help", false, "show usage")
var ollamaRunner = flag.Bool("ollama", false, "uses ollama llm runner")
var openaiRunner = flag.Bool("openai", false, "uses openai llm runner")
var postgres = flag.Bool("postgres", false, "uses postgres as vector store")
var dolt = flag.Bool("dolt", false, "uses dolt as vector store")
var model = flag.String("model", "", "the LLM model to use")
var mariadb = flag.Bool("mariadb", false, "uses mariadb as vector store")
var prompt = flag.String("prompt", "", "the prompt to run")
var docType = flag.String("doc-type", "", "the type of document you are storing")
var storeName = flag.String("store-name", "", "the name of the vector store to use")
var host = flag.String("host", "", "the vector store host to connect to")
var port = flag.Int("port", 0, "the vector store port to connect to")
var user = flag.String("user", "", "the vector store user to connect to")
var topic = flag.String("topic", "", "the topic of the content to generate")
var length = flag.Int("length", 500, "the length of the content to generate")
var vectorDimensions = flag.Int("vector-dimensions", 1536, "the number of dimensions to use for the vector store")
var outputFormat = flag.String("output-format", "markdown", "the output format of the content to generate")
var includeFileExt = flag.String("include-file-ext", ".md", "the file extension used to filter files which should be included in the store")

func main() {
	flag.Parse()

	if *help {
		Usage()
		return
	}

	if *host == "" || *port == 0 {
		printErrorUsageAndExit(errors.New("host and port are required"))
	}
	if *user == "" {
		printErrorUsageAndExit(errors.New("user is required"))
	}
	if *model == "" {
		printErrorUsageAndExit(errors.New("model is required"))
	}

	storePassword := os.Getenv("VECTOR_STORE_PASSWORD")

	docsInputsDir := os.Getenv("DOCS_DIR")

	var runner pkg.Runner
	var sn pkg.StoreType

	if *ollamaRunner {
		runner = pkg.OllamaRunner
	} else if *openaiRunner {
		if os.Getenv("OPENAI_API_KEY") == "" {
			printErrorUsageAndExit(errors.New("OPENAI_API_KEY is required"))
		}
		runner = pkg.OpenAIRunner
	} else {
		printErrorUsageAndExit(errors.New("unsupported runner"))
	}

	if *postgres {
		sn = pkg.Postgres
	} else if *dolt {
		sn = pkg.Dolt
	} else if *mariadb {
		if *vectorDimensions == 0 {
			printErrorUsageAndExit(errors.New("vector dimensions are required for mariadb"))
		}
		sn = pkg.MariaDB
	} else {
		printErrorUsageAndExit(errors.New("unsupported store"))
	}

	if *storeName == "" {
		printErrorUsageAndExit(errors.New("store name is required"))
	}

	storeOnly := false
	var sourceType pkg.DocSourceType
	var splitter textsplitter.TextSplitter

	if *docType != "" {
		if *includeFileExt == "" {
			printErrorUsageAndExit(errors.New("include-file-ext is required"))
		}

		storeOnly = true
		sourceType = pkg.DocSourceType(*docType)

		// todo: make this configurable
		splitter = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithModelName(*model),
			textsplitter.WithChunkSize(512),    // default is 512
			textsplitter.WithChunkOverlap(128), // default is 100
			textsplitter.WithCodeBlocks(true),
			textsplitter.WithHeadingHierarchy(true),
			textsplitter.WithCodeBlocks(true),
		)

	} else {
		splitter = pkg.NewNoopTextSplitter()
	}

	includeFileFunc := func(path string) bool {
		return filepath.Ext(path) == *includeFileExt
	}

	if !storeOnly {
		if *topic == "" {
			printErrorUsageAndExit(errors.New("topic is required"))
		}
		if *length == 0 {
			printErrorUsageAndExit(errors.New("length is required"))
		}
	} else {
		if _, err := os.Stat(docsInputsDir); os.IsNotExist(err) {
			printErrorUsageAndExit(errors.New("docs input dir does not exist"))
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

	config := pkg.NewConfig()
	config.WithRunner(runner)
	config.WithModel(pkg.Model(*model))
	config.WithStoreType(sn)
	config.WithHost(*host)
	config.WithUser(*user)
	config.WithPassword(storePassword)
	config.WithPort(*port)
	config.WithVectorDimensions(*vectorDimensions)
	config.WithStoreName(*storeName)
	config.WithSplitter(splitter)
	config.WithIncludeFileFunc(includeFileFunc)
	config.WithPreContentSystemPrompt(SystemPromptPreContentBlock)
	config.WithPostContentSystemPrompt(SystemPromptPostContentBlock)

	blogger, err := pkg.NewBlogger(
		ctx,
		config,
		logger,
	)
	if err != nil {
		printErrorUsageAndExit(err)
	}
	defer blogger.Close()

	if storeOnly {
		err = blogger.Store(ctx, sourceType, docsInputsDir)
	} else {
		err = blogger.Generate(ctx, *prompt, *topic, *length, *outputFormat)
	}
	if err != nil {
		printErrorUsageAndExit(err)
	}
}

func Usage() {
	fmt.Println("robot-blogger [options]")
	flag.PrintDefaults()
}

func printErrorUsageAndExit(err error) {
	fmt.Println(err)
	Usage()
	os.Exit(1)
}
