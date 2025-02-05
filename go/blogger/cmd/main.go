package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()

	// vectorize inputs mode
	// this mode will vectorize the inputs and save the vectorized inputs to the database

	// start model server
	// defer stop model server

	// start database server
	// defer stop database server

	// read from database the last vectorized input
	// search the provide inputs

	// if the provided inputs are newer than the last vectorized input, then vectorize the inputs
	// and save the vectorized inputs to the database
	// update the last vectorized input in the database

	// if the provided inputs are older than the last vectorized input, then do nothing
	// think we just need to figure out the right key for inputs

	// in query mode

	// use the RAG process to generate a response
	// this will read from the database to get the whatever,
	// then send that to the model server to get a response

	// print the response
	fmt.Println("Hello, World!")
}

func usage() {
	fmt.Println("Usage: blogger <command> [options]")
	flag.PrintDefaults()
}
