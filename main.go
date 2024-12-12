package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/mymcpserver/tools"
)

func main() {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"MyMCP",
		"1.0.0",
		server.WithLogging(),
	)

	tools.RegisterWebSearchTool(mcpServer)
	tools.RegisterFetchTool(mcpServer)
	tools.RegisterConfluenceTool(mcpServer)
	tools.RegisterYouTubeTool(mcpServer)
	tools.RegisterJiraTool(mcpServer)
	tools.RegisterGitLabTool(mcpServer)
	tools.RegisterExpertTool(mcpServer)

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
