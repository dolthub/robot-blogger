package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/zap"
)

var help = flag.Bool("help", false, "show usage")
var ollamaRunner = flag.Bool("ollama", false, "uses ollama llm runner")
var openaiRunner = flag.Bool("openai", false, "uses openai llm runner")
var llama3Model = flag.Bool("llama3", true, "uses the llama3 model for generating the content")
var chatgpt4oModel = flag.Bool("chatgpt4o", false, "uses the chatgpt-4o-latest model for generating the content")
var postgres = flag.Bool("postgres", false, "uses postgres as vector store")
var dolt = flag.Bool("dolt", false, "uses dolt as vector store")
var mariadb = flag.Bool("mariadb", false, "uses mariadb as vector store")
var prompt = flag.String("prompt", "", "the prompt to run")
var storeBlogs = flag.Bool("store-blogs", false, "store dolthub blog documents")
var storeEmails = flag.Bool("store-emails", false, "store dolthub marketing email documents")
var storeCustom = flag.String("store-custom", "", "store custom documents")
var storeName = flag.String("store-name", "", "the name of the vector store to use")
var host = flag.String("host", "", "the host to connect to")
var port = flag.Int("port", 0, "the port to connect to")
var user = flag.String("user", "", "the user of the vector store")
var topic = flag.String("topic", "", "the topic of the content to generate")
var length = flag.Int("length", 500, "the length of the content to generate")
var vectorDimensions = flag.Int("vector-dimensions", 1536, "the number of dimensions to use for the vector store")
var outputFormat = flag.String("output-format", "markdown", "the output format of the content to generate")

func main() {
	flag.Parse()

	if *help {
		Usage()
		return
	}

	if *host == "" || *port == 0 {
		panic("host and port are required")
	}
	if *user == "" {
		panic("user is required")
	}

	storePassword := os.Getenv("VECTOR_STORE_PASSWORD")

	dolthubBlogInputsDir := os.Getenv("DOLTHUB_BLOGS_DIR")
	// dolthubEmailsInputsDir := os.Getenv("DOLTHUB_EMAILS_DIR")

	var runner Runner
	var model Model
	var sn StoreType

	if *ollamaRunner {
		runner = OllamaRunner
	} else if *openaiRunner {
		if os.Getenv("OPENAI_API_KEY") == "" {
			panic("OPENAI_API_KEY is required")
		}
		runner = OpenAIRunner
	} else {
		panic("unsupported runner")
	}

	if *llama3Model {
		model = Llama3Model
	} else if *chatgpt4oModel {
		model = ChatGPT4oModel
	} else {
		panic("unsupported model")
	}

	if *postgres {
		sn = Postgres
	} else if *dolt {
		sn = Dolt
	} else if *mariadb {
		if *vectorDimensions == 0 {
			panic("vector dimensions are required for mariadb")
		}
		sn = MariaDB
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
			textsplitter.WithModelName(string(model)),
			textsplitter.WithChunkSize(512),    // default is 512
			textsplitter.WithChunkOverlap(128), // default is 100
			textsplitter.WithCodeBlocks(true),
			textsplitter.WithHeadingHierarchy(true),
			textsplitter.WithCodeBlocks(true),
		)

		if _, err := os.Stat(dolthubBlogInputsDir); os.IsNotExist(err) {
			panic("dolthub blog inputs dir does not exist")
		}

		inputsDir = dolthubBlogInputsDir

		includeFileFunc = func(path string) bool {
			return filepath.Ext(path) == ".md"
		}

	} else if *storeEmails {
		storeOnly = true

		panic("not implemented")

	} else if *storeCustom != "" {
		storeOnly = true

		panic("not implemented")
	} else {
		splitter = NewNoopTextSplitter()
	}

	if !storeOnly {
		if *topic == "" {
			panic("topic is required")
		}
		if *length == 0 {
			panic("length is required")
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

	blogger, err := NewBlogger(
		ctx,
		runner,
		model,
		sn,
		*host,
		*user,
		storePassword,
		*port,
		*vectorDimensions,
		*storeName,
		splitter,
		includeFileFunc,
		DocSourceTypeBlogPost,
		logger,
	)
	if err != nil {
		panic(err)
	}

	if storeOnly {
		err = blogger.Store(ctx, inputsDir)
	} else {
		err = blogger.Generate(ctx, *prompt, *topic, *length, *outputFormat)
	}
	if err != nil {
		panic(err)
	}
}

func Usage() {
	fmt.Println("robot-blogger [options]")
	flag.PrintDefaults()
}
