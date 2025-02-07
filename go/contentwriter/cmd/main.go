package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	postgres2 "github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs/postgres"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/modelrunner/ollama"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer/llama3"
	"go.uber.org/zap"
)

var model = flag.String("model", "llama3", "the model to use for generating the content")
var raw = flag.Bool("raw", false, "use raw model, no embeddings db")
var query = flag.String("query", "", "the query to run")
var inputsDir = flag.String("inputs", "", "the inputs directory")
var postgres = flag.Bool("postgres", false, "use postgres for the database")

// todo: fix args
// --llama3 --query
// --llama3 --postgres --query
// --llama3 -inputs <dir>

func main() {
	flag.Parse()

	if *model == "" {
		fmt.Println("model is required")
		usage()
		os.Exit(1)
	}

	prompt := WriteDoltBlogPostInMarkdownPromptNoEmbeddings_v1
	if *query != "" {
		prompt = *query
	}

	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error initializing logger:", err)
		os.Exit(1)
	}

	// db is noop for now
	var db dbs.DatabaseServer
	if *postgres {
		db, err = postgres2.NewPostgresLocallyRunningServer(ctx, logger)
	} else {
		db = dbs.NewNoopDatabaseServer()
	}
	if err != nil {
		fmt.Println("error initializing db:", err)
		os.Exit(1)
	}

	start := time.Now()
	defer func() {
		logger.Info("content writer total time", zap.Duration("duration", time.Since(start)))
	}()

	if *inputsDir != "" {
		err := embedLlama3Inputs(ctx, *inputsDir, *model, db, logger)
		if err != nil {
			fmt.Println("error embedding inputs", err)
			os.Exit(1)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("error getting cwd:", err)
			os.Exit(1)
		}
		today := time.Now().Format("2006-01-02")
		f, err := os.Create(filepath.Join(cwd, fmt.Sprintf("%s-%s-generated.md", today, *model)))
		if err != nil {
			fmt.Println("error creating blog file:", err)
			os.Exit(1)
		}
		defer f.Close()

		if *raw {
			err = writeRawLlama3Content(ctx, *model, prompt, f, logger)
		} else {
			err = writeRAGLlama3Content(ctx, *model, prompt, db, f, logger)
		}

		if err != nil {
			fmt.Println("error writing blog", err)
			os.Exit(1)
		}
	}
}

func embedLlama3Inputs(ctx context.Context, inputsDir string, model string, db dbs.DatabaseServer, logger *zap.Logger) error {
	inputs, err := writer.NewMarkdownBlogPostInputsFromDir(inputsDir)
	if err != nil {
		return err
	}

	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, logger)
	if err != nil {
		return err
	}

	cw, err := llama3.NewLlama3(ctx, modelServer, db, logger)
	if err != nil {
		return err
	}
	defer cw.Close(ctx)

	for _, input := range inputs {
		err = cw.UpdateInput(ctx, input)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeRawLlama3Content(ctx context.Context, model, prompt string, wc io.WriteCloser, logger *zap.Logger) error {
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, logger)
	if err != nil {
		return err
	}
	db := dbs.NewNoopDatabaseServer()
	rawBlogger, err := llama3.NewLlama3(ctx, modelServer, db, logger)
	if err != nil {
		return err
	}
	defer rawBlogger.Close(ctx)

	_, err = rawBlogger.WriteContent(ctx, prompt, wc)
	return err
}

func writeRAGLlama3Content(ctx context.Context, model, prompt string, db dbs.DatabaseServer, wc io.WriteCloser, logger *zap.Logger) error {
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, logger)
	if err != nil {
		return err
	}
	embedBlogger, err := llama3.NewLlama3(ctx, modelServer, db, logger)
	if err != nil {
		return err
	}
	defer embedBlogger.Close(ctx)

	_, err = embedBlogger.WriteContent(ctx, prompt, wc)
	return err
}

func usage() {
	fmt.Println("Usage: content writer <command> [options]")
	flag.PrintDefaults()
}
