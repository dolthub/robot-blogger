package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs"
	dolt2 "github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs/dolt"
	postgres2 "github.com/dolthub/robot-blogger/go/contentwriter/pkg/dbs/postgres"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/modelrunner/ollama"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer"
	"github.com/dolthub/robot-blogger/go/contentwriter/pkg/writer/llama3"
	"go.uber.org/zap"
)

var llama3Model = flag.Bool("llama3", true, "uses the llama3 model for generating the content")
var postgres = flag.Bool("postgres", false, "uses postgres to store embeddings")
var dolt = flag.Bool("dolt", false, "uses dolt to store embeddings")
var query = flag.String("query", "", "the query to run")
var inputsDir = flag.String("inputs", "", "the inputs directory")

func main() {
	flag.Parse()

	model := writer.Llama3

	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error initializing logger:", err)
		os.Exit(1)
	}

	var db dbs.DatabaseServer
	if *postgres {
		db, err = postgres2.NewPostgresServer(ctx, logger)
	} else if *dolt {
		db, err = dolt2.NewDoltServer(ctx, logger)
	} else {
		db = dbs.NewNoopDatabaseServer()
	}

	prompt := WriteDoltBlogPostInMarkdownPromptNoEmbeddings_v1
	if *query != "" {
		prompt = *query
	}

	start := time.Now()
	defer func() {
		logger.Info("content writer total time", zap.Duration("duration", time.Since(start)))
	}()

	if *inputsDir != "" {
		err := embedInputs(ctx, *inputsDir, model, db, logger)
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
		f, err := os.Create(filepath.Join(cwd, fmt.Sprintf("%s-%s-generated.md", today, model)))
		if err != nil {
			fmt.Println("error creating blog file:", err)
			os.Exit(1)
		}
		defer f.Close()

		err = writeContent(ctx, prompt, model, db, f, logger)

		if err != nil {
			fmt.Println("error writing blog", err)
			os.Exit(1)
		}
	}
}

func embedInputs(ctx context.Context, inputsDir string, model writer.ModelName, db dbs.DatabaseServer, logger *zap.Logger) error {
	inputs, err := writer.NewMarkdownBlogPostInputsFromDir(inputsDir)
	if err != nil {
		return err
	}

	switch model {
	case writer.Llama3:
		mr, err := ollama.NewOllamaModelRunner(string(model), logger)
		if err != nil {
			return err
		}
		l3, err := llama3.NewLlama3(ctx, mr, db, logger)
		if err != nil {
			return err
		}
		defer l3.Close(ctx)
		for _, input := range inputs {
			err = l3.UpdateInput(ctx, input)
			if err != nil {
				return err
			}
		}
		if db.Name() == dbs.Mysql {
			err = l3.CreateIndex(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("unsupported model: %s", model)
}

func writeContent(ctx context.Context, prompt string, model writer.ModelName, db dbs.DatabaseServer, wc io.WriteCloser, logger *zap.Logger) error {
	switch model {
	case writer.Llama3:
		mr, err := ollama.NewOllamaModelRunner(string(model), logger)
		if err != nil {
			return err
		}
		l3, err := llama3.NewLlama3(ctx, mr, db, logger)
		if err != nil {
			return err
		}
		defer l3.Close(ctx)
		_, err = l3.WriteContent(ctx, prompt, wc)
		return err
	}

	return fmt.Errorf("unsupported model: %s", model)
}

func usage() {
	fmt.Println("Usage: contentwriter <command> [options]")
	flag.PrintDefaults()
}
