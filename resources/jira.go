package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
)

func RegisterJiraResource(s *server.MCPServer) {
	template := mcp.NewResourceTemplate(
		"jira://{id}",
		"Jira Issue",
		mcp.WithTemplateDescription("Returns details of a Jira issue"),
		mcp.WithTemplateMIMEType("text/markdown"),
		mcp.WithTemplateAnnotations([]mcp.Role{mcp.RoleAssistant, mcp.RoleUser}, 0.5),
	)

	// Add resource with its handler
	s.AddResourceTemplate(template, func(request mcp.ReadResourceRequest) ([]interface{}, error) {

		requestURI := request.Params.URI
		issueKey := strings.TrimPrefix(requestURI, "jira://")
		client := services.JiraClient()

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

		// Get reporter name, handling nil case
		reporterName := "Unassigned"
		if issue.Fields.Reporter != nil {
			reporterName = issue.Fields.Reporter.DisplayName
		}

		// Get assignee name, handling nil case
		assigneeName := "Unassigned"
		if issue.Fields.Assignee != nil {
			assigneeName = issue.Fields.Assignee.DisplayName
		}

		// Get priority name, handling nil case
		priorityName := "None"
		if issue.Fields.Priority != nil {
			priorityName = issue.Fields.Priority.Name
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
			reporterName,
			assigneeName,
			issue.Fields.Created,
			issue.Fields.Updated,
			priorityName,
			issue.Fields.Description,
			subtasks,
			transitions,
		)

		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      "jira://" + issueKey,
					MIMEType: "text/markdown",
				},
				Text: string(result),
			},
		}, nil
	})
}
