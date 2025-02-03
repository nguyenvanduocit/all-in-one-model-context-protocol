package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"github.com/sashabaranov/go-openai"
)

var envFilePath string

func RegisterToolManagerTool(s *server.MCPServer, envFile string) {
	envFilePath = envFile

	tool := mcp.NewTool("tool_manager",
		mcp.WithDescription("Manage MCP tools - enable or disable tools"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: list, enable, disable")),
		mcp.WithString("tool_name", mcp.Description("Tool name to enable/disable")),
	)

	s.AddTool(tool, util.ErrorGuard(toolManagerHandler))

	planTool := mcp.NewTool("tool_use_plan",
		mcp.WithDescription("Tạo kế hoạch sử dụng các công cụ đang kích hoạt để giải quyết yêu cầu"),
		mcp.WithString("request", mcp.Required(), mcp.Description("Yêu cầu cần lập kế hoạch")),
		mcp.WithString("context", mcp.Required(), mcp.Description("Ngữ cảnh liên quan đến yêu cầu")),
	)
	s.AddTool(planTool, util.ErrorGuard(toolUsePlanHandler))
}

func toolManagerHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	action, ok := arguments["action"].(string)
	if !ok {
		return mcp.NewToolResultError("action must be a string"), nil
	}

	env, err := godotenv.Read(envFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read env file %s: %v", envFilePath, err)), nil
	}

	enableTools := env["ENABLE_TOOLS"]
	toolList := strings.Split(enableTools, ",")

	switch action {
	case "list":
		response := fmt.Sprintf("Environment file: %s\n\n", envFilePath)
		
		allEnabled := enableTools == ""
		
		// List all available tools with status
		response += "Available tools:\n"
		tools := []struct {
			name string
			desc string
		}{
			{"tool_manager", "Tool management"},
			{"gemini", "AI tools: web search"},
			{"fetch", "Web content fetching"},
			{"confluence", "Confluence integration"},
			{"youtube", "YouTube transcript"},
			{"jira", "Jira issue management"},
			{"gitlab", "GitLab integration"},
			{"script", "Script execution"},
			{"rag", "RAG memory tools"},
			{"gmail", "Gmail tools"},
			{"calendar", "Google Calendar tools"},
			{"youtube_channel", "YouTube channel tools"},
			{"sequential_thinking", "Sequential thinking tool"},
			{"deepseek", "Deepseek reasoning tool"},
		}

		for _, t := range tools {
			status := "disabled"
			if allEnabled || contains(toolList, t.name) {
				status = "enabled"
			}
			response += fmt.Sprintf("- %s (%s) [%s]\n", t.name, t.desc, status)
		}
		response += "\n"

		// List enabled tools
		response += "Currently enabled tools:\n"
		if allEnabled {
			response += "All tools are enabled (ENABLE_TOOLS is empty)\n"
		} else {
			for _, tool := range toolList {
				if tool != "" {
					response += fmt.Sprintf("- %s\n", tool)
				}
			}
		}
		return mcp.NewToolResultText(response), nil

	case "enable", "disable":
		toolName, ok := arguments["tool_name"].(string)
		if !ok || toolName == "" {
			return mcp.NewToolResultError("tool_name is required for enable/disable actions"), nil
		}

		// Nếu ENABLE_TOOLS trống, tạo một list mới
		if enableTools == "" {
			toolList = []string{}
		}

		if action == "enable" {
			// Kiểm tra xem tool đã được enable chưa
			if !contains(toolList, toolName) {
				toolList = append(toolList, toolName)
			}
		} else {
			// Disable tool bằng cách xóa khỏi list
			toolList = removeString(toolList, toolName)
		}

		// Cập nhật ENABLE_TOOLS trong env
		env["ENABLE_TOOLS"] = strings.Join(toolList, ",")

		// Ghi lại vào file .env
		content := ""
		for key, value := range env {
			content += fmt.Sprintf("%s=%s\n", key, value)
		}
		if err := os.WriteFile(envFilePath, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write env file %s: %v", envFilePath, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully %sd tool: %s in %s", action, toolName, envFilePath)), nil

	default:
		return mcp.NewToolResultError("Invalid action. Use 'list', 'enable', or 'disable'"), nil
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeString(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func toolUsePlanHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	request, _ := arguments["request"].(string)
	contextString, _ := arguments["context"].(string)

	env, err := godotenv.Read(envFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read env file: %v", err)), nil
	}

	enabledTools := strings.Split(env["ENABLE_TOOLS"], ",")
	if !contains(enabledTools, "deepseek") {
		return mcp.NewToolResultError("Deepseek tool must be enabled to generate plans"), nil
	}

	// Create a more detailed system prompt
	systemPrompt := fmt.Sprintf(`You are a tool usage planning assistant. Create a detailed execution plan using the currently enabled tools: %s

Context: %s

Output format:
1. [Tool Name] - Purpose: ... (Expected result: ...)
2. [Tool Name] - Purpose: ... (Expected result: ...)
...`, strings.Join(enabledTools, ", "), contextString)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: request,
		},
	}

	// Sử dụng client chung từ deepseek.go
	resp, err := deepseekClient().CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       "deepseek-reasoner", // Sử dụng model R1
			Messages:    messages,
			Temperature: 0.3, // Giảm temperature để kế hoạch ổn định hơn
		},
	)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("API call failed: %v", err)), nil
	}

	if len(resp.Choices) == 0 {
		return mcp.NewToolResultError("No response from Deepseek"), nil
	}

	// Format kết quả
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	return mcp.NewToolResultText("📝 **Execution Plan:**\n" + content), nil
} 