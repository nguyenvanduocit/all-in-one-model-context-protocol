package tools

import (
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

func RegisterFetchTool(s *server.MCPServer) {
	tool := mcp.NewTool("fetch_url",
		mcp.WithDescription("Fetch/read a http URL and return the content"),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL to fetch")),
	)

	s.AddTool(tool, util.ErrorGuard(fetchHandler))
}

func fetchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	url, ok := arguments["url"].(string)
	if !ok {
		return mcp.NewToolResultError("url must be a string"), nil
	}

	resp, err := services.DefaultHttpClient().Get(url)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to fetch URL: %s", err)), nil
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read response body: %s", err)), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
