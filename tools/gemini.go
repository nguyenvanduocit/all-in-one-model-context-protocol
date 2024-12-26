package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"google.golang.org/genai"
)

func RegisterGeminiTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("ai_web_search",
		mcp.WithDescription("search the web by using Google AI Search. Best tool to update realtime information"),
		mcp.WithString("question", mcp.Required(), mcp.Description("The question to ask. Should be a question")),
		// context
		mcp.WithString("context", mcp.Required(), mcp.Description("Context/purpose of the question, helps Gemini to understand the question better")),
	)

	s.AddTool(searchTool, util.ErrorGuard(aiWebSearchHandler))

	criticalThinkingTool := mcp.NewTool("ai_critical_thinking",
		mcp.WithDescription("critical thinking tool"),
		mcp.WithString("question", mcp.Required(), mcp.Description("The question to ask. Should be a question")),
		mcp.WithString("context", mcp.Required(), mcp.Description("Context/purpose of the question, helps Gemini to understand the question better")),
	)

	s.AddTool(criticalThinkingTool, util.ErrorGuard(aiCriticalThinkingHandler))
}

var genAiClient = sync.OnceValue[*genai.Client](func() *genai.Client {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		panic("GOOGLE_AI_API_KEY environment variable must be set")
	}

	cfg := &genai.ClientConfig{
		APIKey: apiKey,
		Backend:  genai.BackendGoogleAI,
	}

	client, err := genai.NewClient(context.Background(), cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create Gemini client: %s", err))
	}

	return client
})

func aiCriticalThinkingHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	question, ok := arguments["question"].(string)
	if !ok {
		return mcp.NewToolResultError("question must be a string"), nil
	}

	systemInstruction := "You are an AI assistant capable of critical thinking. Analyze the question step-by-step using scientific reasoning. Consider different perspectives, identify potential biases, and evaluate the evidence before arriving at a well-reasoned conclusion. Explain your thought process clearly and concisely."

	questionContext, ok := arguments["context"].(string)
	if !ok {
		systemInstruction += "\n\nContext: " + questionContext
	}

	resp, err := genAiClient().Models.GenerateContent(context.Background(),
		"gemini-2.0-flash-thinking-exp-1219",
		genai.PartSlice{
			genai.Text(question),
		},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.Text(systemInstruction).ToContent(),
		},
	)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate content: %s", err)), nil
	}

	if len(resp.Candidates) == 0 {
		return mcp.NewToolResultError("no response from Gemini"), nil
	}

	candidate := resp.Candidates[0]

	var textBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		textBuilder.WriteString(part.Text)
	}

	if candidate.CitationMetadata != nil {
		for _, citation := range candidate.CitationMetadata.Citations {
			textBuilder.WriteString("\n\nSource: ")
			textBuilder.WriteString(citation.URI)
		}
	}

	if candidate.GroundingMetadata != nil {
		textBuilder.WriteString("\n\nSources: ")
		for _, chunk := range candidate.GroundingMetadata.GroundingChunks {
			if chunk.RetrievedContext != nil {
				textBuilder.WriteString("\n")
				textBuilder.WriteString(chunk.RetrievedContext.Text)
				textBuilder.WriteString(": ")
				textBuilder.WriteString(chunk.RetrievedContext.URI)
			}

			if chunk.Web != nil {
				textBuilder.WriteString("\n")
				textBuilder.WriteString(chunk.Web.Title)
				textBuilder.WriteString(": ")
				textBuilder.WriteString(chunk.Web.URI)
			}
		}
	}

	return mcp.NewToolResultText(textBuilder.String()), nil
}

func aiWebSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	question, ok := arguments["question"].(string)
	if !ok {
		return mcp.NewToolResultError("question must be a string"), nil
	}

	systemInstruction := "You are a search engine. You will search the web for the answer to the question. You will then provide the answer to the question. Always try to search the web for the answer first before providing the answer. writing style: short, concise, direct, and to the point."

	questionContext, ok := arguments["context"].(string)
	if !ok {
		systemInstruction += "\n\nContext: " + questionContext
	}

	resp, err := genAiClient().Models.GenerateContent(context.Background(),
		"gemini-2.0-flash-exp",
		genai.PartSlice{
			genai.Text(question),
		},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.Text(systemInstruction).ToContent(),
			Tools: []*genai.Tool{
				{GoogleSearch: &genai.GoogleSearch{}},
			},
		},
	)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate content: %s", err)), nil
	}

	if len(resp.Candidates) == 0 {
		return mcp.NewToolResultError("no response from Gemini"), nil
	}

	candidate := resp.Candidates[0]

	var textBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		textBuilder.WriteString(part.Text)
	}

	if candidate.CitationMetadata != nil {
		for _, citation := range candidate.CitationMetadata.Citations {
			textBuilder.WriteString("\n\nSource: ")
			textBuilder.WriteString(citation.URI)
		}
	}

	if candidate.GroundingMetadata != nil {
		textBuilder.WriteString("\n\nSources: ")
		for _, chunk := range candidate.GroundingMetadata.GroundingChunks {
			if chunk.RetrievedContext != nil {
				textBuilder.WriteString("\n")
				textBuilder.WriteString(chunk.RetrievedContext.Text)
				textBuilder.WriteString(": ")
				textBuilder.WriteString(chunk.RetrievedContext.URI)
			}

			if chunk.Web != nil {
				textBuilder.WriteString("\n")
				textBuilder.WriteString(chunk.Web.Title)
				textBuilder.WriteString(": ")
				textBuilder.WriteString(chunk.Web.URI)
			}
		}
	}

	return mcp.NewToolResultText(textBuilder.String()), nil
}

