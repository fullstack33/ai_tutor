package orachastrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/fullstack33/ai-tutor/internal/config"
	"github.com/fullstack33/ai-tutor/internal/repository"
	qdrantrepo "github.com/fullstack33/ai-tutor/internal/repository/qdrant"
	"github.com/fullstack33/ai-tutor/use-cases/embedding"
	"github.com/fullstack33/ai-tutor/use-cases/input"
	"github.com/fullstack33/ai-tutor/use-cases/llms"
	"github.com/fullstack33/ai-tutor/use-cases/read-file"
	"github.com/fullstack33/ai-tutor/use-cases/tokenizer"
)

type IOrachastrator interface {
	Handle() error
}

type Orachastrator struct{}

func NewOrachastrator() IOrachastrator {
	return &Orachastrator{}
}

func (o *Orachastrator) Handle() error {
	fmt.Println("=== AI Tutor RAG Orchestrator ===")
	inputHandler := input.NewInputHandler()

	// 1. Ask user for PDF/file path
	filePath, err := inputHandler.Ask("Enter the path to the PDF/text file: ")
	if err != nil {
		return fmt.Errorf("failed to read file path: %w", err)
	}
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Ask user for their preferred Ollama LLM model name
	llmModel, err := inputHandler.Ask(fmt.Sprintf("Enter the local Ollama LLM model to use (default: %s): ", config.DefaultLLMModel))
	if err != nil {
		return fmt.Errorf("failed to read model name: %w", err)
	}
	if llmModel == "" {
		llmModel = config.DefaultLLMModel
	}

	fmt.Println("\n--------------------------------------------------")

	// 2. Read and extract plain text from the file
	fmt.Println("[STEP 1/5] Ingesting and reading document...")
	readFile := readfile.NewReadFile(filePath)
	text, err := readFile.Handle()
	if err != nil {
		fmt.Println("Orachastrator: Got error in reading file", err)
		return err
	}
	fmt.Printf("[SUCCESS] Document parsed successfully! (Extracted %d characters)\n\n", len(text))

	// 3. Tokenize text into chunks
	fmt.Println("[STEP 2/5] Splitting document text into token chunks...")
	tokenizerObj := tokenizer.NewTokenizer(text, 100, 20)
	tokens, err := tokenizerObj.Handle()
	if err != nil {
		fmt.Println("Orachastrator: Got error in tokenizing text", err)
		return err
	}
	fmt.Printf("[SUCCESS] Text chunking complete! Generated %d overlapping chunks.\n\n", len(tokens))

	// 4. Convert chunks into vector form
	fmt.Println("[STEP 3/5] Connecting to Ollama and generating embeddings...")
	points := []repository.VectorPoint{}
	embed := embedding.NewNomicEmbedding(config.OllamaBaseURL, config.OllamaEmbeddingURL)
	for id, token := range tokens {
		vector, err := embed.GetEmbedding(config.EmbeddingModel, token)
		if err != nil {
			fmt.Println("Orachastrator: Got error in vector conversion embedding ", err)
			return err
		}
		points = append(points, repository.VectorPoint{
			ID:     uint64(id),
			Vector: vector.Embedding,
			Payload: map[string]any{
				"text":        token,
				"chunk_index": id,
			},
		})
	}
	fmt.Printf("[SUCCESS] Embeddings generated successfully! (Vector count: %d)\n\n", len(points))

	// 5. Store the data into Vector DB (Qdrant)
	fmt.Println("[STEP 4/5] Initializing Qdrant database repository...")
	ctx := context.Background()
	vectorDB, err := qdrantrepo.NewQdrantRepository(config.QdrantHost, config.QdrantPort)
	if err != nil {
		fmt.Println("Orachastrator: Got error initializing Qdrant repository", err)
		return err
	}
	defer vectorDB.Close()

	collectionName := config.QdrantCollection
	if len(points) > 0 {
		vectorSize := uint64(len(points[0].Vector))
		
		// Clear existing collection first to remove old data
		fmt.Printf("[STEP 5/5] Clearing older data by recreating collection '%s'...\n", collectionName)
		err = vectorDB.DeleteCollection(ctx, collectionName)
		if err != nil {
			fmt.Println("Orachastrator: Got error clearing old collection", err)
			return err
		}

		fmt.Printf("Creating Qdrant collection '%s' (Dimensions: %d) and uploading vectors...\n", collectionName, vectorSize)
		err = vectorDB.CreateCollection(ctx, collectionName, vectorSize)
		if err != nil {
			fmt.Println("Orachastrator: Got error creating collection", err)
			return err
		}

		err = vectorDB.UpsertVectors(ctx, collectionName, points)
		if err != nil {
			fmt.Println("Orachastrator: Got error upserting vectors", err)
			return err
		}
		fmt.Printf("[SUCCESS] Vectors uploaded and indexed successfully in Qdrant!\n")
	}
	fmt.Printf("--------------------------------------------------\n\n")

	// Initialize LLM Client
	llmClient := llms.NewOllamaLLM(config.OllamaBaseURL)

	// 6. Start the interactive conversation loop
	fmt.Println("=== Ready! Ask questions about your document below (type 'exit' or 'quit' to end) ===")
	for {
		query, err := inputHandler.Ask("\nYou: ")
		if err != nil {
			fmt.Println("Error reading query:", err)
			continue
		}

		// Handle quit condition
		cleanedQuery := strings.ToLower(strings.TrimSpace(query))
		if cleanedQuery == "exit" || cleanedQuery == "quit" {
			fmt.Println("Exiting conversational interface. Goodbye!")
			break
		}
		if cleanedQuery == "" {
			continue
		}

		// A. Generate embedding for user's query
		fmt.Println("[QUERY] Generating embedding for your question...")
		queryEmbedding, err := embed.GetEmbedding(config.EmbeddingModel, query)
		if err != nil {
			fmt.Println("Error generating embedding for your query:", err)
			continue
		}

		// B. Query Qdrant for top 3 relevant context chunks
		fmt.Println("[SEARCH] Querying Qdrant vector database for top 3 matching segments...")
		qdrantResults, err := vectorDB.QueryVectors(ctx, collectionName, queryEmbedding.Embedding, 3)
		if err != nil {
			fmt.Println("Error searching Qdrant database:", err)
			continue
		}
		fmt.Printf("[SUCCESS] Retrieved %d context segments from Qdrant!\n", len(qdrantResults))

		// C. Concat top 3 responses
		var contextBuilder strings.Builder
		for i, res := range qdrantResults {
			textVal, ok := res.Point.Payload["text"].(string)
			if ok {
				contextBuilder.WriteString(fmt.Sprintf("[Context Segment %d]\n%s\n\n", i+1, textVal))
			}
		}
		contextText := contextBuilder.String()

		// D. Construct RAG Prompt
		ragPrompt := fmt.Sprintf(`You are a helpful AI Tutor. Use the following context retrieved from the user's uploaded document to answer their question as accurately and informatively as possible. If the answer is not supported by the context, state that clearly.

Context:
%s

Question: %s

Answer:`, contextText, query)

		// E. Call the local LLM via Ollama
		fmt.Printf("[LLM] Sending context prompt to local Ollama LLM (%s)...\n", llmModel)
		response, err := llmClient.Generate(llmModel, ragPrompt)
		if err != nil {
			fmt.Println("Error calling LLM:", err)
			continue
		}
		fmt.Println("[SUCCESS] Tutor response received!")

		// F. Display response to user
		fmt.Println("\nAI Tutor Response:")
		fmt.Println(response)
	}

	return nil
}
