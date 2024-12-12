package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/mymcpserver/services"
)

// RegisterJiraTool registers the Jira tools to the server
func RegisterJiraTool(s *server.MCPServer) {
	// Add Jira tool
	jiraTool := mcp.NewTool("get_jira_issue",
		mcp.WithDescription("Get Jira issue details"),
		mcp.WithString("issue_key", mcp.Required(), mcp.Description("Jira issue key (e.g., KP-2)")),
	)

	s.AddTool(jiraTool, jiraIssueHandler)
}

func jiraIssueHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	client := services.JiraClient()

	// Get issue key from arguments
	issueKey, ok := arguments["issue_key"].(string)
	if !ok {
		return nil, fmt.Errorf("issue_key argument is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	issue, response, err := client.Issue.Get(ctx, issueKey, nil, []string{"transitions"})
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get issue: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get issue: %v", err)
	}

	// Build subtasks string if they exist
	var subtasks string
	if issue.Fields.Subtasks != nil {
		subtasks = "\nSubtasks:\n"
		for _, subTask := range issue.Fields.Subtasks {
			subtasks += fmt.Sprintf("- %s: %s\n", subTask.Key, subTask.Fields.Summary)
		}
	}

	// Build transitions string
	var transitions string
	for _, transition := range issue.Transitions {
		transitions += fmt.Sprintf("- %s (ID: %s)\n", transition.Name, transition.ID)
	}

	result := fmt.Sprintf(`
Key: %s
Summary: %s
Status: %s
Reporter: %s
Assignee: %s
Created: %s
Updated: %s
Priority: %s
Description:
%s
%s
Available Transitions:
%s`,
		issue.Key,
		issue.Fields.Summary,
		issue.Fields.Status.Name,
		issue.Fields.Reporter.DisplayName,
		issue.Fields.Assignee.DisplayName,
		issue.Fields.Created,
		issue.Fields.Updated,
		issue.Fields.Priority.Name,
		issue.Fields.Description,
		subtasks,
		transitions,
	)

	return mcp.NewToolResultText(result), nil
}
