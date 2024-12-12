package tools

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/option"
)

func RegisterCommandLineExpertTool(s *server.MCPServer) {
	tool := mcp.NewTool("command_line_expert",
		mcp.WithDescription("Use Gemini to get command line suggestions"),
		mcp.WithString("request", mcp.Required(), mcp.Description("The command line request")),
	)

	s.AddTool(tool, commandLineExpertHandler)
}

var genAiClient = sync.OnceValue[*genai.Client](func() *genai.Client {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		panic("GOOGLE_AI_API_KEY environment variable must be set")
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		panic(fmt.Sprintf("failed to create Gemini client: %s", err))
	}

	return client
})

func commandLineExpertHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	request, ok := arguments["request"].(string)
	if !ok {
		return mcp.NewToolResultError("request must be a string"), nil
	}

	model := genAiClient().GenerativeModel("gemini-2.0-flash-exp")
	resp, err := model.GenerateContent(context.Background(), genai.Text(request))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate content: %s", err)), nil
	}

	if len(resp.Candidates) == 0 {
		return mcp.NewToolResultError("no response from Gemini"), nil
	}

	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	return mcp.NewToolResultText(text), nil
}
