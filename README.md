# robot-blogger
a robot blog writer


## my notes

following this blog: https://pgdash.io/blog/rag-with-postgresql.html

Steps:
1. download and install ollama, https://github.com/ollama/ollama/blob/main/README.md#quickstart
2. install postrgres and pgvector (if installing via homebrew, need postgresql@17), https://github.com/pgvector/pgvector
3. create a new database called `robot_blogger`
3. run create extension vector in postgres