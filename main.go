package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/prompts"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/resources"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/tools"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	flag.Parse()

	if err := godotenv.Load(*envFile); err != nil {
		fmt.Printf("Warning: Error loading env file %s: %v\n", *envFile, err)
	}
	mcpServer := server.NewMCPServer(
		"MyMCP",
		"1.0.0",
		server.WithLogging(),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	enableTools := strings.Split(os.Getenv("ENABLE_TOOLS"), ",")
	allToolsEnabled := len(enableTools) == 1 && enableTools[0] == ""

	isEnabled := func(toolName string) bool {
		return allToolsEnabled || slices.Contains(enableTools, toolName)
	}

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

	if isEnabled("youtube_channel") {
		tools.RegisterYouTubeChannelTools(mcpServer)
	}

	prompts.RegisterCodeTools(mcpServer)

	resources.RegisterJiraResource(mcpServer)

	if err := server.ServeStdio(mcpServer); err != nil {
		panic(fmt.Sprintf("Server error: %v", err))
	}
}
