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
	"github.com/sashabaranov/go-openai"
)

func RegisterDeepseekTool(s *server.MCPServer) {
	reasoningTool := mcp.NewTool("deepseek_reasoning",
		mcp.WithDescription("Deepseek-powered reasoning tool for complex problem solving and analysis"),
		mcp.WithString("question", mcp.Required(), mcp.Description("The question or problem to analyze")),
		mcp.WithString("context", mcp.Required(), mcp.Description("Context/purpose of the question, helps Deepseek to understand the problem better")),
		mcp.WithString("chat_context_or_knowledge_base_summary", mcp.Description("Current chat history/knowledge base summary to inform the analysis")),
	)

	s.AddTool(reasoningTool, util.ErrorGuard(deepseekReasoningHandler))
}

var deepseekClient = sync.OnceValue(func() *openai.Client {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		panic("DEEPSEEK_API_KEY environment variable must be set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"

	client := openai.NewClientWithConfig(config)
	return client
})

func deepseekReasoningHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	question, ok := arguments["question"].(string)
	if !ok {
		return mcp.NewToolResultError("question must be a string"), nil
	}

	contextArgument, ok := arguments["context"].(string)
	if !ok {
		contextArgument = ""
	}

	chatContext, _ := arguments["chat_context"].(string)

	systemPrompt := `You are an advanced reasoning engine powered by Deepseek. Your task is to:
1. Break down complex problems into manageable components
2. Apply systematic reasoning and logical analysis
3. Consider multiple perspectives and potential implications
4. Identify assumptions and potential biases
5. Draw well-supported conclusions based on available information
6. Provide clear explanations of your reasoning process

Context for this analysis: ` + contextArgument

	if chatContext != "" {
		systemPrompt += "\n\nAdditional Chat Context:\n" + chatContext
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: question,
		},
	}

	resp, err := deepseekClient().CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       "deepseek-reasoner",
			Messages:    messages,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate content: %s", err)), nil
	}

	if len(resp.Choices) == 0 {
		return mcp.NewToolResultError("no response from Deepseek"), nil
	}

	var textBuilder strings.Builder
	textBuilder.WriteString(resp.Choices[0].Message.Content)

	return mcp.NewToolResultText(textBuilder.String()), nil
} 