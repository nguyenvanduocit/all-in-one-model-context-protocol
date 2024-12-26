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
	tool := mcp.NewTool("get_url_content",
		mcp.WithDescription("Fetches content from a given HTTP/HTTPS URL. This tool allows you to retrieve text content from web pages, APIs, or any accessible HTTP endpoints. Returns the raw content as text."),
		mcp.WithString("url", 
			mcp.Required(), 
			mcp.Description("The complete HTTP/HTTPS URL to fetch content from (e.g., https://example.com)"),
		),
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
