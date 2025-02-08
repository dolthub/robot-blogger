package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("llama3"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := generateFromVectors(ctx, llm, "What is Dolt and DoltHub?")
	//completion, err := generateFromSinglePrompt(ctx, llm, "Human: Who was the first man to walk on the moon?\nAssistant:")
	if err != nil {
		log.Fatal(err)
	}

	_ = completion
}

func generateFromVectors(ctx context.Context, llm *ollama.LLM, prompt string) (string, error) {
	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	dir := "/Users/dustin/src/ld/web/packages/blog/src/pages"

	files := make([]string, 0)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	// todo: remove this
	// files = files[:1]

	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(512),    // default is 512
		textsplitter.WithChunkOverlap(128), // default is 100
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithHeadingHierarchy(true),
	)

	url := fmt.Sprintf("postgres://%s@%s:%d/%s", "postgres", "127.0.0.1", 5432, "robot_blogger_llama3_v4")
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(url),
		pgvector.WithEmbedder(e),
	)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		docs, err := textsplitter.CreateDocuments(splitter, []string{string(content)}, nil)
		if err != nil {
			return "", err
		}

		start := time.Now()
		fmt.Println("embedding docs for blog post:", filepath.Base(file))
		_, err = store.AddDocuments(ctx, docs)
		if err != nil {
			return "", err
		}
		fmt.Println("done embedding docs for blog post:", filepath.Base(file), time.Since(start))
	}

	docs, err := store.SimilaritySearch(ctx, prompt, 10)
	//fmt.Println(docs)

	//// Prompt Template
	//promptTemplate := prompts.NewPromptTemplate("Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.\n\n{context}\n\nQuestion: {question}\nHelpful Answer:", []string{"context", "question"})
	//
	//// Chain
	//chain := chains.NewLLMChain(llm, promptTemplate)
	//
	//// Call chain with query and document contents
	//result, err := chains.Call(ctx, chain, map[string]any{
	//	"context":  docs,
	//	"question": prompt,
	//})
	//if err != nil {
	//	return "", err
	//}
	//fmt.Println(result["text"].(string))
	//
	//return result["text"].(string), nil

	fullPrompt := prompt
	if len(docs) > 0 {
		fullPrompt = "Use the following pieces of context to answer the question at the end. The context pieces are as follows:\n"
		for idx, doc := range docs {
			fullPrompt += "context piece " + strconv.Itoa(idx+1) + ": \n"
			fullPrompt += fmt.Sprintf("%s\n", doc.PageContent)
			fullPrompt += "end of context piece " + strconv.Itoa(idx+1) + "\n\n"
		}
		fullPrompt += "The question is: " + prompt + "\n\n"
	}

	// this kinda works
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		fullPrompt,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		return "", err
	}

	return completion, nil
}

func generateFromSinglePrompt(ctx context.Context, llm *ollama.LLM, prompt string) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		prompt,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		return "", err
	}

	return completion, nil
}
