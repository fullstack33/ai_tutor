## Description
- RAG implementation for AI Tutor

## Tech Stack
- Golang
- lancedb (Vector Db)[https://pkg.go.dev/github.com/lancedb/lancedb-go/pkg/lancedb]
- qdrant [https://qdrant.tech/documentation/quickstart/]
- nomic-embed-text [https://ollama.com/library/nomic-embed-text]
- llama3.1:8b


## Flow
- Read pdf
- Tokenize pdf
    - 1500 words token
    - overlap 200 words
- Convert text token in embedding
- store embedding token in vector db
- Retrive data from vector
- Find the most relavent token data 
- Pass token text data to llm as context