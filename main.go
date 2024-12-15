package main

import (
	"fmt"

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

	// normal search
	//tools.RegisterWebSearchTool(mcpServer)

	// Gemini powered search
	tools.RegisterExpertTool(mcpServer)

	// Fetch tool
	tools.RegisterFetchTool(mcpServer)

	// Confluence tool
	tools.RegisterConfluenceTool(mcpServer)

	// YouTube tool
	tools.RegisterYouTubeTool(mcpServer)

	// Jira tool
	tools.RegisterJiraTool(mcpServer)

	// GitLab tool
	tools.RegisterGitLabTool(mcpServer)

	// CLI tool
	tools.RegisterScriptTool(mcpServer)

	// Vector tool
	tools.RegisterVectorTool(mcpServer)
	

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		panic(fmt.Sprintf("Server error: %v", err))
	}
}
