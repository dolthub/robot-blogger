# robot-blogger

a RAG AI blog/email writer with Dolt as an optional vector store.

# Dependencies

* [golang](https://go.dev/doc/install)
* [ollama](https://ollama.com/download)

... more to come

# Installation

```bash
cd robot-blogger
go install .
```

## Required Flags

**Required Flags**

- `-host`, the vector store host to connect to.
- `-port`, the vector store port to connect to.
- `-user`, the vector store user to connect to.
- `-model`, the LLM model to use.
- `-store-name`, the name of the vector store to use.

One of:

- `-dolt`, uses dolt as vector store.
- `-mariadb`, uses mariadb as vector store.
- `-postgres`, uses postgres as vector store.

One of:

- `-ollama`, uses ollama llm runner. `OLLAMA_HOST` environment variable must be set if not running on localhost.
- `-openai`, uses openai llm runner. `OPENAI_API_KEY` environment variable must be set.

**Required for Store**

- `DOCS_DIR` environment variable must be set. Specifies path to directory containing the docs to store.
- `-doc-type`, the type of document you are storing.
- `-include-file-ext`, the file extension of the docs to store.

**Required for Generate**

- `-prompt`, the prompt to run.
- `-topic`, the topic of the content to generate.
- `-length`, the length of the content to generate.

**Optional Flags**

- `-output-format`, the format of the content to generate.
- `-vector-dimensions`, the number of dimensions to use for the vector store. Required for MariaDB.

You must have a vector store running with the store name/database name already created. You may also need to pull the model
you are trying to use, ie:

```bash
ollama pull llama3
```

## Store

```bash
export VECTOR_STORE_PASSWORD=mydbpass
export DOCS_DIR=/path/to/docs

./robot-blogger \
--ollama \
--model=llama3 \
--dolt \
--user=root \
--host=0.0.0.0 \
--port=3306 \
--store-name=robot_blogger_llama3_v1 \
--doc-type=blog_post \
--include-file-ext=".md"
```

## Generate

```bash
export VECTOR_STORE_PASSWORD=mydbpass

./robot-blogger \
--ollama \
--model=llama3 \
--dolt \
--user=root \
--host=0.0.0.0 \
--port=3306 \
--topic="DoltHub Products" \
--length=100 \
--store-name=robot_blogger_llama3_v1 \
--prompt="What are Dolt and DoltHub?"
```
