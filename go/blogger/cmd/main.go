package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	postgres2 "github.com/dolthub/robot-blogger/go/blogger/pkg/dbs/postgres"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger/llama3"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models/ollama"
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

	// db is noop for now
	var db dbs.DatabaseServer
	var err error
	if *postgres {
		db, err = postgres2.NewPostgresLocallyRunningServer(ctx)
	} else {
		db = dbs.NewNoopDatabaseServer()
	}
	if err != nil {
		fmt.Println("error initializing db:", err)
		os.Exit(1)
	}

	if *inputsDir != "" {
		err := embedInputs(ctx, *inputsDir, *model, db)
		if err != nil {
			fmt.Println("error embedding inputs", err)
			os.Exit(1)
		}
	} else {
		err := writeBlog(ctx, *model, db, os.Stdout)
		if err != nil {
			fmt.Println("error writing blog", err)
			os.Exit(1)
		}
	}
}

func embedInputs(ctx context.Context, inputsDir string, model string, db dbs.DatabaseServer) error {
	inputs, err := blogger.NewMarkdownBlogPostInputsFromDir(inputsDir)
	if err != nil {
		return err
	}

	// todo: for now, only do single input
	inputs = inputs[:1]

	// todo: move the model server start stop to blogger
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model)
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

func writeBlog(ctx context.Context, model string, db dbs.DatabaseServer, wc io.WriteCloser) error {
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model)
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
