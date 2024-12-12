package tools

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	"gitlab.com/gitlab-org/api/client-go"
)

var gitlabClient = sync.OnceValue[*gitlab.Client](func() *gitlab.Client {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		log.Fatal("GITLAB_TOKEN is required")
	}

	host := os.Getenv("GITLAB_HOST")
	if host == "" {
		log.Fatal("GITLAB_HOST is required")
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(host))
	if err != nil {
		log.Fatal(errors.WithMessage(err, "failed to create gitlab client"))
	}

	return client
})

// RegisterGitLabTool registers the GitLab tool with the MCP server
func RegisterGitLabTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("gitlab_search_projects",
		mcp.WithDescription("Search GitLab projects"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	)

	projectTool := mcp.NewTool("gitlab_get_project",
		mcp.WithDescription("Get GitLab project details"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
	)

	mrListTool := mcp.NewTool("gitlab_list_mrs",
		mcp.WithDescription("List merge requests"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("state", mcp.DefaultString("opened"), mcp.Description("MR state (opened/closed/merged)")),
	)

	mrDetailsTool := mcp.NewTool("gitlab_get_mr_details",
		mcp.WithDescription("Get merge request details"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	mrCommentTool := mcp.NewTool("gitlab_create_MR_note",
		mcp.WithDescription("Create a note on a merge request"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithString("comment", mcp.Required(), mcp.Description("Comment text")),
	)

	mrChangesTool := mcp.NewTool("gitlab_list_mr_changes",
		mcp.WithDescription("Get detailed changes of a merge request"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	fileContentTool := mcp.NewTool("gitlab_get_file_content",
		mcp.WithDescription("Get file content from a GitLab repository"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file in the repository")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch name, tag, or commit SHA")),
	)

	pipelineTool := mcp.NewTool("gitlab_list_pipelines",
		mcp.WithDescription("List pipelines for a GitLab project"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("status", mcp.DefaultString("all"), mcp.Description("Pipeline status (running/pending/success/failed/canceled/skipped/all)")),
	)

	s.AddTool(searchTool, searchProjectsHandler)
	s.AddTool(projectTool, getProjectHandler)
	s.AddTool(mrListTool, listMergeRequestsHandler)
	s.AddTool(mrDetailsTool, getMergeRequestHandler)
	s.AddTool(mrCommentTool, commentOnMergeRequestHandler)
	s.AddTool(mrChangesTool, getMergeRequestChangesHandler)
	s.AddTool(fileContentTool, getFileContentHandler)
	s.AddTool(pipelineTool, listPipelinesHandler)
}

func searchProjectsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	query := arguments["query"].(string)

	opt := &gitlab.ListProjectsOptions{
		Search: &query,
	}

	projects, _, err := gitlabClient().Projects.ListProjects(opt)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %v", err)
	}

	var result string
	for _, project := range projects {
		result += fmt.Sprintf("ID: %d\nName: %s\nPath: %s\nDescription: %s\nURL: %s\n\n",
			project.ID, project.Name, project.PathWithNamespace, project.Description, project.WebURL)
	}

	return mcp.NewToolResultText(result), nil
}

func getProjectHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)

	// Get project details
	project, _, err := gitlabClient().Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	// Get branches
	branches, _, err := gitlabClient().Branches.ListBranches(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %v", err)
	}

	// Get tags
	tags, _, err := gitlabClient().Tags.ListTags(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %v", err)
	}

	// Build basic project info
	result := fmt.Sprintf("Project Details:\nID: %d\nName: %s\nPath: %s\nDescription: %s\nURL: %s\nDefault Branch: %s\n\n",
		project.ID, project.Name, project.PathWithNamespace, project.Description, project.WebURL,
		project.DefaultBranch)

	// Add branches
	result += "Branches:\n"
	for _, branch := range branches {
		result += fmt.Sprintf("- %s\n", branch.Name)
	}

	// Add tags
	result += "\nTags:\n"
	for _, tag := range tags {
		result += fmt.Sprintf("- %s\n", tag.Name)
	}

	return mcp.NewToolResultText(result), nil
}

func listMergeRequestsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	state := arguments["state"].(string)

	opt := &gitlab.ListProjectMergeRequestsOptions{
		State: gitlab.String(state),
	}

	mrs, _, err := gitlabClient().MergeRequests.ListProjectMergeRequests(projectID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list merge requests: %v", err)
	}

	var result string
	for _, mr := range mrs {
		result += fmt.Sprintf("MR #%d: %s\nState: %s\nAuthor: %s\nURL: %s\nCreated: %s\n\n",
			mr.IID, mr.Title, mr.State, mr.Author.Username, mr.WebURL, mr.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

func getMergeRequestHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	mrIIDStr := arguments["mr_iid"].(string)

	mrIID, err := strconv.Atoi(mrIIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid mr_iid: %v", err)
	}

	mr, _, err := gitlabClient().MergeRequests.GetMergeRequest(projectID, mrIID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge request: %v", err)
	}

	// Get MR changes
	changes, _, err := gitlabClient().MergeRequests.ListMergeRequestDiffs(projectID, mrIID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge request changes: %v", err)
	}

	result := fmt.Sprintf("MR Details:\nTitle: %s\nState: %s\nAuthor: %s\nURL: %s\nCreated: %s\nDescription:\n%s\n\nChanges:\n",
		mr.Title, mr.State, mr.Author.Username, mr.WebURL, mr.CreatedAt.Format("2006-01-02 15:04:05"), mr.Description)

	for _, change := range changes {
		switch change.NewFile {
		case true:
			result += fmt.Sprintf("\nFile: %s\nStatus: Added\n", change.NewPath)
		case false:
			result += fmt.Sprintf("\nFile: %s\nStatus: Deleted\n", change.OldPath)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func commentOnMergeRequestHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	mrIIDStr := arguments["mr_iid"].(string)
	comment := arguments["comment"].(string)

	mrIID, err := strconv.Atoi(mrIIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid mr_iid: %v", err)
	}

	opt := &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.String(comment),
	}

	note, _, err := gitlabClient().Notes.CreateMergeRequestNote(projectID, mrIID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	result := fmt.Sprintf("Comment posted successfully!\nID: %d\nAuthor: %s\nCreated: %s\nContent: %s",
		note.ID, note.Author.Username, note.CreatedAt.Format("2006-01-02 15:04:05"), note.Body)

	return mcp.NewToolResultText(result), nil
}

func getMergeRequestChangesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	mrIIDStr := arguments["mr_iid"].(string)

	mrIID, err := strconv.Atoi(mrIIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid mr_iid: %v", err)
	}

	// Get MR details first
	mr, _, err := gitlabClient().MergeRequests.GetMergeRequest(projectID, mrIID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge request: %v", err)
	}

	// Get detailed changes
	changes, _, err := gitlabClient().MergeRequests.ListMergeRequestDiffs(projectID, mrIID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge request changes: %v", err)
	}

	var result strings.Builder

	// Write MR overview
	result.WriteString(fmt.Sprintf("Merge Request #%d: %s\n", mr.IID, mr.Title))
	result.WriteString(fmt.Sprintf("Author: %s\n", mr.Author.Username))
	result.WriteString(fmt.Sprintf("Source Branch: %s\n", mr.SourceBranch))
	result.WriteString(fmt.Sprintf("Target Branch: %s\n", mr.TargetBranch))
	result.WriteString(fmt.Sprintf("State: %s\n", mr.State))
	result.WriteString(fmt.Sprintf("Created: %s\n", mr.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("URL: %s\n", mr.WebURL))
	// Add SHAs information
	result.WriteString(fmt.Sprintf("Base SHA: %s\n", mr.DiffRefs.BaseSha))
	result.WriteString(fmt.Sprintf("Start SHA: %s\n", mr.DiffRefs.StartSha))
	result.WriteString(fmt.Sprintf("Head SHA: %s\n\n", mr.DiffRefs.HeadSha))

	if mr.Description != "" {
		result.WriteString("Description:\n")
		result.WriteString(mr.Description)
		result.WriteString("\n\n")
	}

	// Write changes overview
	result.WriteString(fmt.Sprintf("Changes Overview:\n"))
	result.WriteString(fmt.Sprintf("Total files changed: %d\n\n", len(changes)))

	// Write detailed changes for each file
	for _, change := range changes {
		result.WriteString(fmt.Sprintf("File: %s\n", change.NewPath))
		result.WriteString(fmt.Sprintf("Status: %s\n", getChangeStatus(change)))

		if change.Diff != "" {
			result.WriteString("Diff:\n")
			result.WriteString("```diff\n")
			result.WriteString(change.Diff)
			result.WriteString("\n```\n")
		}

		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

// Helper function to determine file change status
func getChangeStatus(change *gitlab.MergeRequestDiff) string {
	if change.NewFile {
		return "Added"
	}
	if change.DeletedFile {
		return "Deleted"
	}
	if change.RenamedFile {
		return fmt.Sprintf("Renamed from %s", change.OldPath)
	}
	return "Modified"
}

func getFileContentHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	filePath := arguments["file_path"].(string)
	ref := arguments["ref"].(string)

	// Get raw file content
	fileContent, _, err := gitlabClient().RepositoryFiles.GetRawFile(projectID, filePath, &gitlab.GetRawFileOptions{
		Ref: gitlab.Ptr(ref),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %v; maybe wrong ref?", err)
	}

	var result strings.Builder

	// Write file information
	result.WriteString(fmt.Sprintf("File: %s\n", filePath))
	result.WriteString(fmt.Sprintf("Ref: %s\n", ref))
	result.WriteString("Content:\n")
	result.WriteString(string(fileContent))

	return mcp.NewToolResultText(result.String()), nil
}

func listPipelinesHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	status := arguments["status"].(string)

	opt := &gitlab.ListProjectPipelinesOptions{}
	if status != "all" {
		opt.Status = gitlab.Ptr(gitlab.BuildStateValue(status))
	}

	pipelines, _, err := gitlabClient().Pipelines.ListProjectPipelines(projectID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Pipelines for project %s:\n\n", projectID))

	for _, pipeline := range pipelines {
		result.WriteString(fmt.Sprintf("Pipeline #%d\n", pipeline.ID))
		result.WriteString(fmt.Sprintf("Status: %s\n", pipeline.Status))
		result.WriteString(fmt.Sprintf("Ref: %s\n", pipeline.Ref))
		result.WriteString(fmt.Sprintf("SHA: %s\n", pipeline.SHA))
		result.WriteString(fmt.Sprintf("Created: %s\n", pipeline.CreatedAt.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("URL: %s\n\n", pipeline.WebURL))
	}

	return mcp.NewToolResultText(result.String()), nil
}
