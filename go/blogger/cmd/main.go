package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/models"
)

var model = flag.String("model", "llama3", "the model to use for generating the blog")
var port = flag.Int("port", 11434, "the port to use for the model server")
var createEmbeddings = flag.Bool("create-embeddings", false, "creates embeddings with the model and writes them to the database")
var prompt = flag.String("prompt", "", "the prompt to use for generating the blog")

func main() {
	flag.Parse()

	if *model == "" {
		fmt.Println("model is required")
		usage()
		os.Exit(1)
	}
	if *port == 0 {
		fmt.Println("port is required")
		usage()
		os.Exit(1)
	}

	ctx := context.Background()

	modelServer := models.NewOllamaLocallyRunningServer(*model, *port)
	err := modelServer.Start(ctx)
	if err != nil {
		fmt.Println("error starting model server", err)
		os.Exit(1)
	}
	defer modelServer.Stop(ctx)

	if *createEmbeddings {
		// start database server
		// defer stop database server

		// read from database the last vectorized input
		// search the provide inputs

		// if the provided inputs are newer than the last vectorized input, then vectorize the inputs
		// and save the vectorized inputs to the database
		// update the last vectorized input in the database

		// if the provided inputs are older than the last vectorized input, then do nothing
		// think we just need to figure out the right key for inputs
	} else if *prompt != "" {

		// in query mode

		// use the RAG process to generate a response
		// this will read from the database to get the whatever,
		// then send that to the model server to get a response

		// print the response
	} else {
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: blogger <command> [options]")
	flag.PrintDefaults()
}
