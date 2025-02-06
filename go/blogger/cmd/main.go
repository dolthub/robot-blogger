package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	postgres2 "github.com/dolthub/robot-blogger/go/blogger/pkg/dbs/postgres"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger/llama3"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models/ollama"
	"go.uber.org/zap"
)

var model = flag.String("model", "llama3", "the model to use for generating the blog")
var inputsDir = flag.String("inputs", "", "the inputs directory")
var postgres = flag.Bool("postgres", false, "use postgres for the database")

func main() {
	flag.Parse()

	if *model == "" {
		fmt.Println("model is required")
		usage()
		os.Exit(1)
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
		logger.Info("robot blogger total time", zap.Duration("duration", time.Since(start)))
	}()

	if *inputsDir != "" {
		err := embedInputs(ctx, *inputsDir, *model, db, logger)
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
		err = writeBlog(ctx, *model, db, f, logger)
		if err != nil {
			fmt.Println("error writing blog", err)
			os.Exit(1)
		}
	}
}

func embedInputs(ctx context.Context, inputsDir string, model string, db dbs.DatabaseServer, logger *zap.Logger) error {
	inputs, err := blogger.NewMarkdownBlogPostInputsFromDir(inputsDir)
	if err != nil {
		return err
	}

	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, logger)
	if err != nil {
		return err
	}

	blgr, err := llama3.NewLlama3BloggerWithEmbeddings(ctx, modelServer, db)
	if err != nil {
		return err
	}
	defer blgr.Close(ctx)

	for _, input := range inputs {
		err = blgr.UpdateInput(ctx, input)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeBlog(ctx context.Context, model string, db dbs.DatabaseServer, wc io.WriteCloser, logger *zap.Logger) error {
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, logger)
	if err != nil {
		return err
	}
	rawBlogger, err := llama3.NewLlama3OnlyBlogger(ctx, modelServer)
	if err != nil {
		return err
	}
	defer rawBlogger.Close(ctx)

	_, err = rawBlogger.WriteBlog(ctx, WriteDoltMarketingStatementPromptNoEmbeddings, wc)
	return err
}

func usage() {
	fmt.Println("Usage: blogger <command> [options]")
	flag.PrintDefaults()
}
