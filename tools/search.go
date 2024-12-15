package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	htmltomarkdownnnn "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"github.com/tidwall/gjson"
)

func RegisterWebSearchTool(s *server.MCPServer) {
	tool := mcp.NewTool("web_search",
		mcp.WithDescription("Search the web using Brave Search API"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Query to search for (max 400 chars, 50 words)")),
		mcp.WithNumber("count", mcp.DefaultNumber(5), mcp.Description("Number of results (1-20, default 5)")),
		mcp.WithString("country", mcp.DefaultString("ALL"), mcp.Description("Country code")),
	)

	s.AddTool(tool, util.ErrorGuard(webSearchHandler))
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Age         string `json:"age"`
}

func webSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	query, ok := arguments["query"].(string)
	if !ok {
		return mcp.NewToolResultError("query must be a string"), nil
	}

	count := 10
	if countArg, ok := arguments["count"].(float64); ok {
		count = int(countArg)
		if count < 1 {
			count = 1
		} else if count > 20 {
			count = 20
		}
	}

	country := "ALL"
	if countryArg, ok := arguments["country"].(string); ok {
		country = countryArg
	}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return mcp.NewToolResultError("BRAVE_API_KEY environment variable is required"), nil
	}

	baseURL := "https://api.search.brave.com/res/v1/web/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("count", fmt.Sprintf("%d", count))
	params.Add("country", country)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	resp, err := services.DefaultHttpClient().Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to perform search: %v", err)), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read response: %v", err)), nil
	}

	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("API request failed: %s", string(body))), nil
	}

	var results []*SearchResult

	gbody := gjson.ParseBytes(body)

	videoResults := gbody.Get("videos.results")
	for _, video := range videoResults.Array() {

		mdContent, err := htmltomarkdownnnn.ConvertString(video.Get("description").String())
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTML to Markdown: %v", err)
		}

		results = append(results, &SearchResult{
			Title:       video.Get("title").String(),
			URL:         video.Get("url").String(),
			Description: mdContent,
			Type:        "video",
			Age:         video.Get("age").String(),
		})
	}

	webResults := gbody.Get("web.results")
	for _, web := range webResults.Array() {

		mdContent, err := htmltomarkdownnnn.ConvertString(web.Get("description").String())
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTML to Markdown: %v", err)
		}

		results = append(results, &SearchResult{
			Title:       web.Get("title").String(),
			URL:         web.Get("url").String(),
			Description: mdContent,
			Type:        "web",
			Age:         web.Get("age").String(),
		})
	}

	if len(results) == 0 {
		return mcp.NewToolResultError("No results found, pls try again with a different query"), nil
	}

	responseText := ""
	for _, result := range results {
		responseText += fmt.Sprintf("Title: %s\nURL: %s\nDescription: %s\nType: %s\nAge: %s\n\n",
			result.Title, result.URL, result.Description, result.Type, result.Age)
	}

	return mcp.NewToolResultText(responseText), nil
}
