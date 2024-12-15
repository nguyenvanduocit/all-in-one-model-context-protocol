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

	enableTools := strings.Split(os.Getenv("ENABLE_TOOLS"), ",")
	allToolsEnabled := len(enableTools) == 1 && enableTools[0] == ""

	// normal search
	//tools.RegisterWebSearchTool(mcpServer)

	// Gemini powered search
	if allToolsEnabled || slices.Contains(enableTools, "gemini") {
		tools.RegisterGeminiTool(mcpServer)
	}

	// Fetch tool
	if allToolsEnabled || slices.Contains(enableTools, "fetch") {
		tools.RegisterFetchTool(mcpServer)
	}

	// Confluence tool
	if allToolsEnabled || slices.Contains(enableTools, "confluence") {
		tools.RegisterConfluenceTool(mcpServer)
	}

	// YouTube tool
	if allToolsEnabled || slices.Contains(enableTools, "youtube") {
		tools.RegisterYouTubeTool(mcpServer)
	}

	// Jira tool
	if allToolsEnabled || slices.Contains(enableTools, "jira") {
		tools.RegisterJiraTool(mcpServer)
	}

	// GitLab tool
	if allToolsEnabled || slices.Contains(enableTools, "gitlab") {
		tools.RegisterGitLabTool(mcpServer)
	}

	// CLI tool
	if allToolsEnabled || slices.Contains(enableTools, "script") {
		tools.RegisterScriptTool(mcpServer)
	}

	// Rag tool
	if allToolsEnabled || slices.Contains(enableTools, "rag") {
		tools.RegisterRagTools(mcpServer)
	}

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		panic(fmt.Sprintf("Server error: %v", err))
	}
}