package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

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
	db := dbs.NewNoopDatabaseServer()

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

	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, db)
	if err != nil {
		return err
	}

	err = modelServer.Start(ctx)
	if err != nil {
		return err
	}
	defer modelServer.Stop(ctx)

	for _, input := range inputs {
		err = modelServer.GenerateEmbeddings(ctx, input.ID())
		if err != nil {
			return err
		}
	}

	return nil
}

func writeBlog(ctx context.Context, model string, db dbs.DatabaseServer, wc io.WriteCloser) error {
	modelServer, err := ollama.NewOllamaLocallyRunningServer(model, db)
	if err != nil {
		return err
	}
	err = modelServer.Start(ctx)
	if err != nil {
		return err
	}
	defer modelServer.Stop(ctx)
	rawBlogger := llama3.NewLlama3OnlyBlogger(modelServer)
	_, err = rawBlogger.WriteBlog(ctx, WriteDoltMarketingStatementPromptNoEmbeddings, wc)
	return err
}

func usage() {
	fmt.Println("Usage: blogger <command> [options]")
	flag.PrintDefaults()
}
