# robot-blogger

a RAG AI blog/email writer with Dolt as an optional vector store.

# Dependencies

* [golang](https://go.dev/doc/install)
* [ollama](https://ollama.com/download)

... more to come

# Installation

```bash
cd robot-blogger/go/cmd
go install .
```

# Example

You must have a vector store running with the store name/database name already created. You may also need to pull the model
you are trying to use, ie:

```bash
ollama pull llama3
```

## Store

To Store Content for later RAG use include the `--store-blogs`, `--store-emails`, or `--store-custom` options with the other required flags.

`--store-blogs` requires `DOLTHUB_BLOGS_DIR` environment variable to be set.
`--store-emails` requires `DOLTHUB_EMAILS_DIR` environment variable to be set.

```bash
export VECTOR_STORE_PASSWORD=mydbpass
export DOLTHUB_BLOGS_DIR=/path/to/dolthub/blogs

./robot-blogger \
--ollama \
--llama3 \
--dolt \
--user=root \
--host=0.0.0.0 \
--port=3306 \
--store-name=robot_blogger_llama3_v1 \
--store-blogs
```

## Generate

To Generate RAG Content include the `--prompt` option with the other required flags.

```bash
export VECTOR_STORE_PASSWORD=mydbpass

./robot-blogger \
--ollama \
--llama3 \
--dolt \
--user=root \
--host=0.0.0.0 \
--port=3306 \
--store-name=robot_blogger_llama3_v1 \
--prompt="What are Dolt and DoltHub?"
```


# TODO

* Add reranking via BM25 + embedding rerankings.
* Dynamically adjust the number of chunks retrieved based on the prompt, ie for 500-800 word content, retrieve 3-5 chunks, for 1000-1500 word content, retrieve 5-7 chunks, etc.


