package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/tools"
)

func main() {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"MyMCP",
		"1.0.0",
		server.WithLogging(),
	)

	// Check environment variable first for backward compatibility
	enableTools := strings.Split(os.Getenv("ENABLE_TOOLS"), ",")
	allToolsEnabled := len(enableTools) == 1 && enableTools[0] == ""

	// Helper function to check if a tool should be enabled
	isEnabled := func(toolName string) bool {
		return allToolsEnabled || slices.Contains(enableTools, toolName)
	}

	// Register tools based on preferences
	if isEnabled("gemini") {
		tools.RegisterGeminiTool(mcpServer)
	}

	if isEnabled("fetch") {
		tools.RegisterFetchTool(mcpServer)
	}

	if isEnabled("confluence") {
		tools.RegisterConfluenceTool(mcpServer)
	}

	if isEnabled("youtube") {
		tools.RegisterYouTubeTool(mcpServer)
	}

	if isEnabled("jira") {
		tools.RegisterJiraTool(mcpServer)
	}

	if isEnabled("gitlab") {
		tools.RegisterGitLabTool(mcpServer)
	}

	if isEnabled("script") {
		tools.RegisterScriptTool(mcpServer)
	}

	if isEnabled("rag") {
		tools.RegisterRagTools(mcpServer)
	}

	if isEnabled("gmail") {
		tools.RegisterGmailTools(mcpServer)
	}

	if isEnabled("calendar") {
		tools.RegisterCalendarTools(mcpServer)
	}

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		panic(fmt.Sprintf("Server error: %v", err))
	}
}