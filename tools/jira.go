package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

// RegisterJiraTool registers the Jira tools to the server
func RegisterJiraTool(s *server.MCPServer) {
	// Add Jira tool
	jiraTool := mcp.NewTool("get_jira_issue",
		mcp.WithDescription("Get Jira issue details"),
		mcp.WithString("issue_key", mcp.Required(), mcp.Description("Jira issue key (e.g., KP-2)")),
	)

	// Add Jira search tool
	jiraSearchTool := mcp.NewTool("search_jira_issue",
		mcp.WithDescription("Search/list for Jira issues by JQL"),
		mcp.WithString("jql", mcp.Required(), mcp.Description("JQL query to search/list for Jira issues")),
	)

	 // Add Jira list sprint tool
    jiraListSprintTool := mcp.NewTool("list_jira_sprints",
        mcp.WithDescription("List all sprints in a Jira project"),
        mcp.WithString("board_id", mcp.Required(), mcp.Description("Jira board ID")),
    )

		jiraCreateIssueTool := mcp.NewTool("create_jira_issue",
		mcp.WithDescription("Create a new Jira issue"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key (e.g., KP)")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Summary of the issue")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the issue")),
		mcp.WithString("issue_type", mcp.Required(), mcp.Description("Type of the issue (e.g., Bug, Task)")),
	)

	jiraUpdateIssueTool := mcp.NewTool("update_jira_issue",
    mcp.WithDescription("Update an existing Jira issue"),
    mcp.WithString("issue_key", mcp.Required(), mcp.Description("Jira issue key (e.g., KP-2)")),
    mcp.WithString("summary", mcp.Description("New summary of the issue")),
    mcp.WithString("description", mcp.Description("New description of the issue")),
    mcp.WithString("status", mcp.Description("New status of the issue")),
)

	s.AddTool(jiraTool, util.ErrorGuard(jiraIssueHandler))
	s.AddTool(jiraSearchTool, util.ErrorGuard(jiraSearchHandler))
	s.AddTool(jiraListSprintTool, util.ErrorGuard(jiraListSprintHandler))
	s.AddTool(jiraCreateIssueTool, util.ErrorGuard(jiraCreateIssueHandler))
	s.AddTool(jiraUpdateIssueTool, util.ErrorGuard(jiraUpdateIssueHandler))
}


func jiraUpdateIssueHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
    client := services.JiraClient()

    issueKey, ok := arguments["issue_key"].(string)
    if !ok {
        return nil, fmt.Errorf("issue_key argument is required")
    }

    // Create update payload
    payload := &models.IssueSchemeV2{
        Fields: &models.IssueFieldsSchemeV2{},
    }

    // Check and add optional fields if provided
    if summary, ok := arguments["summary"].(string); ok && summary != "" {
        payload.Fields.Summary = summary
    }

    if description, ok := arguments["description"].(string); ok && description != "" {
        payload.Fields.Description = description
    }

    if status, ok := arguments["status"].(string); ok && status != "" {
        payload.Fields.Status = &models.StatusScheme{
            Name: status,
        }
    }

    ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
    defer cancel()

    response, err := client.Issue.Update(ctx, issueKey, true, payload, nil, nil)
    if err != nil {
        if response != nil {
            return nil, fmt.Errorf("failed to update issue: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
        }
        return nil, fmt.Errorf("failed to update issue: %v", err)
    }

    return mcp.NewToolResultText("Issue updated successfully!"), nil
}

func jiraCreateIssueHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	client := services.JiraClient()

	projectKey, ok := arguments["project_key"].(string)
	if !ok {
		return nil, fmt.Errorf("project_key argument is required")
	}

	summary, ok := arguments["summary"].(string)
	if !ok {
		return nil, fmt.Errorf("summary argument is required")
	}

	description, ok := arguments["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description argument is required")
	}

	issueType, ok := arguments["issue_type"].(string)
	if !ok {
		return nil, fmt.Errorf("issue_type argument is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	var payload = models.IssueSchemeV2{
		Fields: &models.IssueFieldsSchemeV2{
			Summary:   summary,
			Project:   &models.ProjectScheme{Key: projectKey},
			Description: description,
			IssueType: &models.IssueTypeScheme{Name: issueType},
		},
	}

	
	issue, response, err := client.Issue.Create(ctx, &payload, nil)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to create issue: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to create issue: %v", err)
	}

	result := fmt.Sprintf("Issue created successfully!\nKey: %s\nID: %s\nURL: %s", issue.Key, issue.ID, issue.Self)
	return mcp.NewToolResultText(result), nil
}

func jiraListSprintHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {

    boardIDStr, ok := arguments["board_id"].(string)
    if !ok {
        return nil, fmt.Errorf("board_id argument is required")
    }

    boardID, err := strconv.Atoi(boardIDStr)
    if err != nil {
        return nil, fmt.Errorf("invalid board_id: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
    defer cancel()

    sprints, response, err := services.AgileClient().Board.Sprints(ctx, boardID, 0, 50, []string{"active", "future"})
    if err != nil {
        if response != nil {
            return nil, fmt.Errorf("failed to get sprints: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
        }
        return nil, fmt.Errorf("failed to get sprints: %v", err)
    }

    if len(sprints.Values) == 0 {
        return mcp.NewToolResultText("No sprints found for this board."), nil
    }

    var result string
    for _, sprint := range sprints.Values {
        result += fmt.Sprintf("ID: %d\nName: %s\nState: %s\nStartDate: %s\nEndDate: %s\n\n", sprint.ID, sprint.Name, sprint.State, sprint.StartDate, sprint.EndDate)
    }

    return mcp.NewToolResultText(result), nil
}

func jiraSearchHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	client := services.JiraClient()

	// Get search text from arguments
	jql, ok := arguments["jql"].(string)
	if !ok {
		return nil, fmt.Errorf("jql argument is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	searchResult, response, err := client.Issue.Search.Get(ctx, jql, nil, nil, 0, 30, "")
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to search issues: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to search issues: %v", err)
	}

	if len(searchResult.Issues) == 0 {
		return mcp.NewToolResultText("No issues found matching the search criteria."), nil
	}

	var sb strings.Builder
	for _, issue := range searchResult.Issues {
		sb.WriteString(fmt.Sprintf("Key: %s\n", issue.Key))

		if issue.Fields.Summary != "" {
			sb.WriteString(fmt.Sprintf("Summary: %s\n", issue.Fields.Summary))
		}

		if issue.Fields.Status != nil && issue.Fields.Status.Name != "" {
			sb.WriteString(fmt.Sprintf("Status: %s\n", issue.Fields.Status.Name))
		}

		if issue.Fields.Created != "" {
			sb.WriteString(fmt.Sprintf("Created: %s\n", issue.Fields.Created))
		}

		if issue.Fields.Updated != "" {
			sb.WriteString(fmt.Sprintf("Updated: %s\n", issue.Fields.Updated))
		}

		if issue.Fields.Assignee != nil {
			sb.WriteString(fmt.Sprintf("Assignee: %s\n", issue.Fields.Assignee.DisplayName))
		} else {
			sb.WriteString("Assignee: Unassigned\n")
		}

		if issue.Fields.Priority != nil {
			sb.WriteString(fmt.Sprintf("Priority: %s\n", issue.Fields.Priority.Name))
		} else {
			sb.WriteString("Priority: Unset\n") 
		}

		if issue.Fields.Resolutiondate != "" {
			sb.WriteString(fmt.Sprintf("Resolution date: %s\n", issue.Fields.Resolutiondate))
		}

		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
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

	return mcp.NewToolResultText(result), nil
}
