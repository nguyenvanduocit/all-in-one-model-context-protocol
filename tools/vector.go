package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"github.com/pkoukk/tiktoken-go"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
)

var qdrantClient = sync.OnceValue[*qdrant.Client](func() *qdrant.Client {

	host := os.Getenv("QDRANT_HOST")
	if host == "" {
		panic("QDRANT_HOST is not set")
	}

	port := os.Getenv("QDRANT_PORT")
	if port == "" {
		panic("QDRANT_PORT is not set")
	}
	
	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic(fmt.Sprintf("failed to parse QDRANT_PORT: %v", err))
	}

	apiKey := os.Getenv("QDRANT_API_KEY")
	if apiKey == "" {
		panic("QDRANT_API_KEY is not set")
	}
	
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   host,
		Port:   portInt,
		APIKey: apiKey,
		UseTLS: true,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to connect to Qdrant: %v", err))
	}

	return client
})

func RegisterVectorTool(s *server.MCPServer) {
	indexContentTool := mcp.NewTool("RAG_memory_index_content",
		mcp.WithDescription("Index a note into memory, can be inserted or updated"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("note file path")),
		mcp.WithString("payload", mcp.Required(), mcp.Description("Plain text payload")),
	)

	indexFileTool := mcp.NewTool("RAG_memory_index_file",
		mcp.WithDescription("Index a local file into memory"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("Path to the local file to be indexed")),
	)

	createCollectionTool := mcp.NewTool("RAG_memory_create_collection",
		mcp.WithDescription("Create a new vector collection in memory"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
	)

	deleteCollectionTool := mcp.NewTool("RAG_memory_delete_collection",
		mcp.WithDescription("Delete a vector collection in memory"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
	)

	listCollectionTool := mcp.NewTool("RAG_memory_list_collections",
		mcp.WithDescription("List all vector collections in memory"),
	)

	searchTool := mcp.NewTool("RAG_memory_search",
		mcp.WithDescription("Search for notes in a collection based on a query"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	)

	deleteIndexByFilePathTool := mcp.NewTool("RAG_memory_delete_index_by_filepath",
		mcp.WithDescription("Delete a vector index by filePath"),
		mcp.WithString("collection", mcp.Required(), mcp.Description("Memory collection name")),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("Path to the local file to be deleted")),
	)

	s.AddTool(createCollectionTool, util.ErrorGuard(createCollectionHandler))
	s.AddTool(deleteCollectionTool, util.ErrorGuard(deleteCollectionHandler))
	s.AddTool(listCollectionTool, util.ErrorGuard(listCollectionHandler))
	s.AddTool(indexContentTool, util.ErrorGuard(indexContentHandler))
	s.AddTool(searchTool, util.ErrorGuard(vectorSearchHandler))
	s.AddTool(indexFileTool, util.ErrorGuard(indexFileHandler))
	s.AddTool(deleteIndexByFilePathTool, util.ErrorGuard(deleteIndexByFilePathHandler))
}

func deleteIndexByFilePathHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	collection := arguments["collection"].(string)
	filePath := arguments["filePath"].(string)
	ctx := context.Background()

	// Delete points by IDs using PointSelector
	pointsSelector := &qdrant.PointsSelector{
		PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
			Filter: &qdrant.Filter{
				Must: []*qdrant.Condition{
					{
						ConditionOneOf: &qdrant.Condition_Field{
							Field: &qdrant.FieldCondition{
								Key: "filePath",
								Match: &qdrant.Match{
									MatchValue: &qdrant.Match_Text{
										Text: filePath,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	deleteResp, err := qdrantClient().Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collection,
		Points: pointsSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete points for filePath %s: %v", filePath, err)
	}

	result := fmt.Sprintf("Successfully deleted points for filePath: %s\nOperation ID: %d\nStatus: %s", filePath, deleteResp.OperationId, deleteResp.Status)
	return mcp.NewToolResultText(result), nil
}

func indexFileHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	collection := arguments["collection"].(string)
	filePath := arguments["filePath"].(string)

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Prepare arguments for vectorUpsertHandler
	upsertArgs := map[string]interface{}{
		"collection": collection,
		"filePath":   filePath,
		"payload":    string(content), // Convert content to string
	}

	// Call vectorUpsertHandler
	return indexContentHandler(upsertArgs)
}

func listCollectionHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	ctx := context.Background()
	collections, err := qdrantClient().ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return mcp.NewToolResultText(fmt.Sprintf("Collections: %v", collections)), nil
}

func createCollectionHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    collection := arguments["collection"].(string)
    ctx := context.Background()

    // Check if collection already exists
    collectionInfo, err := qdrantClient().GetCollectionInfo(ctx, collection)
    if err == nil && collectionInfo != nil {
        return nil, fmt.Errorf("collection %s already exists", collection)
    }

    // Create collection with configuration for text embeddings
    err = qdrantClient().CreateCollection(ctx, &qdrant.CreateCollection{
        CollectionName: collection,
        VectorsConfig: &qdrant.VectorsConfig{
            Config: &qdrant.VectorsConfig_Params{
                Params: &qdrant.VectorParams{
                    Size:     uint64(2048), // OpenAI text-embedding-3-large dimension
                    Distance: qdrant.Distance_Cosine,
                },
            },
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create collection: %v", err)
    }

    result := fmt.Sprintf("Successfully created collection: %s", collection)
    return mcp.NewToolResultText(result), nil
}

func deleteCollectionHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    collection := arguments["collection"].(string)
    ctx := context.Background()

    // Check if collection exists
    collectionInfo, err := qdrantClient().GetCollectionInfo(ctx, collection)
    if err != nil || collectionInfo == nil {
        return nil, fmt.Errorf("collection %s does not exist", collection)
    }

    // Delete collection
    err = qdrantClient().DeleteCollection(ctx, collection)
    if err != nil {
        return nil, fmt.Errorf("failed to delete collection: %v", err)
    }

    result := fmt.Sprintf("Successfully deleted collection: %s", collection)
    return mcp.NewToolResultText(result), nil
}

func indexContentHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	collection := arguments["collection"].(string)
	filePath := arguments["filePath"].(string)
	payload := arguments["payload"].(string)

	// Split content into chunks
	chunks, err := splitIntoChunks(payload, filePath) // Implement chunking logic
	if err != nil {
		return nil, fmt.Errorf("failed to split into chunks: %v", err)
	}

	var points []*qdrant.PointStruct
	for i, chunk := range chunks {
		// Generate embeddings for each chunk
		resp, err := services.DefaultOpenAIClient().CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
			Input: []string{chunk},
			Model: openai.LargeEmbedding3,
			Dimensions: 2048,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings: %v", err)
		}

		// Create point for each chunk
		point := &qdrant.PointStruct{
			Id: qdrant.NewIDUUID(uuid.NewSHA1(uuid.NameSpaceURL, []byte(filePath + strconv.Itoa(i))).String()),
			Vectors: qdrant.NewVectors(resp.Data[0].Embedding...),
			Payload: qdrant.NewValueMap(map[string]any{
				"filePath": filePath,
				"content": chunk,
				"chunkIndex": i,
			}),
		}
		points = append(points, point)
	}

	ctx := context.Background()
	waitUpsert := true

	// Upsert all chunks
	upsertResp, err := qdrantClient().Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Wait:           &waitUpsert,
		Points:         points,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert points: %v", err)
	}

	result := fmt.Sprintf("Successfully upserted\nOperation ID: %d\nStatus: %s", upsertResp.OperationId, upsertResp.Status)

	return mcp.NewToolResultText(result), nil
}

func splitIntoChunks(content string, filePath string) ([]string, error) {
	const (
		maxTokensPerChunk = 512
		overlapTokens     = 50
		model             = "text-embedding-3-large"
	)

	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding: %v", err)
	}

	tokens := encoding.Encode(content, nil, nil)

	var chunks []string
	var currentChunk []int

	for i := 0; i < len(tokens); i++ {
		currentChunk = append(currentChunk, tokens[i])

		if len(currentChunk) >= maxTokensPerChunk {
			chunkText := encoding.Decode(currentChunk)

			// Generate context for the chunk
			contextualizedChunk, err := generateContext(content, chunkText, filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to generate context: %v", err)
			}

			chunks = append(chunks, contextualizedChunk)

			if len(currentChunk) > overlapTokens {
				currentChunk = currentChunk[len(currentChunk)-overlapTokens:]
			} else {
				currentChunk = []int{}
			}
		}
	}

	if len(currentChunk) > 0 {
		chunkText := encoding.Decode(currentChunk)

		// Generate context for the chunk
		contextualizedChunk, err := generateContext(content, chunkText, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to generate context: %v", err)
		}

		chunks = append(chunks, contextualizedChunk)
	}

	return chunks, nil
}

func generateContext(fullText, chunkText, filePath string) (string, error) {
	prompt := fmt.Sprintf(`
<document>%s</document>
Here is the chunk we want to situate within the whole document:
<chunk>%s</chunk>
Please give a short succinct context to situate this chunk within the overall document for the purposes of improving search retrieval of the chunk. Answer only with the succinct context and nothing else.
File path: %s
	`, fullText, chunkText, filePath)

	resp, err := services.DefaultOpenAIClient().CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oLatest,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate context: %v", err)
	}

	context := resp.Choices[0].Message.Content
	return fmt.Sprintf("Context: %s; Chunk: %s", context, chunkText), nil
}

func vectorSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	collection := arguments["collection"].(string)
	query := arguments["query"].(string)
	// Generate embedding for the query
	resp, err := services.DefaultOpenAIClient().CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{query},
		Model: openai.LargeEmbedding3,
		Dimensions: 2048,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings for query: %v", err)
	}

	// Search Qdrant
	searchResult, err := qdrantClient().Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: collection,
		Query:         qdrant.NewQuery(resp.Data[0].Embedding...),
		WithPayload:    &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search in Qdrant: %v", err)
	}

	// Format results
	var resultText string
	for i, hit := range searchResult {
		content := hit.Payload["content"].GetStringValue()
		filePath := hit.Payload["filePath"].GetStringValue()
		resultText += fmt.Sprintf("Result %d (Score: %f):\nFilePath: %s\nContent: %s\n\n", i+1, hit.Score, filePath, content)
	}

	return mcp.NewToolResultText(resultText), nil
}