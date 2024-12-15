package tools

import (
	"context"
	"fmt"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

// registerConfluenceTool is a function that registers the confluence tools to the server
func RegisterConfluenceTool(s *server.MCPServer) {
	tool := mcp.NewTool("confluence_search",
		mcp.WithDescription("Search Confluence"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Atlassian Confluence Query Language (CQL)")),
	)

	s.AddTool(tool, confluenceSearchHandler)

	// Add new tool for getting page content
	pageTool := mcp.NewTool("get_confluence_page",
		mcp.WithDescription("Get Confluence page content"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
	)
	s.AddTool(pageTool, util.ErrorGuard(confluencePageHandler))
}

// confluenceSearchHandler is a handler for the confluence search tool
func confluenceSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get search query from arguments
	query, ok := arguments["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query argument is required")
	}
	ctx := context.Background()
	options := &models.SearchContentOptions{
		Limit: 5,
	}

	var results string

	contents, response, err := client.Search.Content(ctx, query, options)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("search failed: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}

		return nil, fmt.Errorf("search failed: %v", err)
	}

	// Convert results to map format
	for _, content := range contents.Results {
		results += fmt.Sprintf(`
Title: %s
ID: %s 
Type: %s
Link: %s
Last Modified: %s
Body:
%s
----------------------------------------
`,
			content.Content.Title,
			content.Content.ID,
			content.Content.Type,
			content.Content.Links.Self,
			content.LastModified,
			content.Excerpt,
		)
	}

	return mcp.NewToolResultText(results), nil
}

func confluencePageHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get page ID from arguments
	pageID, ok := arguments["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id argument is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	content, response, err := client.Content.Get(ctx, pageID, []string{"body.storage"}, 1)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get page: %v", err)
	}

	mdContent, err := htmltomarkdown.ConvertString(content.Body.Storage.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to Markdown: %v", err)
	}

	result := fmt.Sprintf(`
Title: %s
ID: %s
Type: %s
Content:
%s
`,
		content.Title,
		content.ID,
		content.Type,
		mdContent,
	)

	return mcp.NewToolResultText(result), nil
}
