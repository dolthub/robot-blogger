package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/blogger/llama3"
	"github.com/dolthub/robot-blogger/go/blogger/pkg/models/ollama"
)

var model = flag.String("model", "llama3", "the model to use for generating the blog")
var inputsDir = flag.String("inputs", "", "the inputs directory")

func main() {
	flag.Parse()

	if *model == "" {
		fmt.Println("model is required")
		usage()
		os.Exit(1)
	}

	ctx := context.Background()

	if *inputsDir != "" {
		inputs, err := blogger.NewMarkdownBlogPostInputsFromDir(*inputsDir)
		if err != nil {
			fmt.Println("error reading inputs", err)
			os.Exit(1)
		}

		for _, input := range inputs {
			fmt.Println("input", input.ID())
		}

		// start database server
		// defer stop database server

		// read from database the last vectorized input
		// search the provide inputs

		// if the provided inputs are newer than the last vectorized input, then vectorize the inputs
		// and save the vectorized inputs to the database
		// update the last vectorized input in the database

		// if the provided inputs are older than the last vectorized input, then do nothing
		// think we just need to figure out the right key for inputs

	} else {
		modelServer, err := ollama.NewOllamaLocallyRunningServer(*model)
		if err != nil {
			fmt.Println("error starting model server", err)
			os.Exit(1)
		}

		err = modelServer.Start(ctx)
		if err != nil {
			fmt.Println("error starting model server", err)
			os.Exit(1)
		}
		defer modelServer.Stop(ctx)
		rawBlogger := llama3.NewLlama3OnlyBlogger(modelServer)
		_, err = rawBlogger.WriteBlog(ctx, WriteDoltMarketingStatementPromptNoEmbeddings, os.Stdout)
		if err != nil {
			fmt.Println("error writing blog", err)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Println("Usage: blogger <command> [options]")
	flag.PrintDefaults()
}
