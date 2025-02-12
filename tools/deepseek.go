package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"github.com/sashabaranov/go-openai"
)

func RegisterDeepseekTool(s *server.MCPServer) {
	reasoningTool := mcp.NewTool("deepseek_reasoning",
		mcp.WithDescription("advanced reasoning engine using Deepseek's AI capabilities for multi-step problem solving, critical analysis, and strategic decision support"),
		mcp.WithString("question", mcp.Required(), mcp.Description("The structured query or problem statement requiring deep analysis and reasoning")),
		mcp.WithString("context", mcp.Required(), mcp.Description("Defines the operational context and purpose of the query within the MCP ecosystem")),
		mcp.WithString("knowledge", mcp.Description("Provides relevant chat history, knowledge base entries, and structured data context for MCP-aware reasoning")),
	)

	s.AddTool(reasoningTool, util.ErrorGuard(deepseekReasoningHandler))
}


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

	systemPrompt := "Context:\n" + contextArgument

	if chatContext != "" {
		systemPrompt += "\n\nAdditional Context:\n" + chatContext
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

	resp, err := services.DefaultDeepseekClient().CreateChatCompletion(
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