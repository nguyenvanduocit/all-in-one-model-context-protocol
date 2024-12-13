package tools

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
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
		mcp.WithString("state", mcp.DefaultString("all"), mcp.Description("MR state (opened/closed/merged)")),
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
		mcp.WithString("ref", mcp.DefaultString("develop"), mcp.Description("Branch name, tag, or commit SHA")),
	)

	pipelineTool := mcp.NewTool("gitlab_list_pipelines",
		mcp.WithDescription("List pipelines for a GitLab project"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("status", mcp.DefaultString("all"), mcp.Description("Pipeline status (running/pending/success/failed/canceled/skipped/all)")),
	)

	commitsTool := mcp.NewTool("gitlab_list_commits",
		mcp.WithDescription("List commits in a GitLab project within a date range"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("since", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
		mcp.WithString("until", mcp.Required(), mcp.Description("End date (YYYY-MM-DD)")),
		mcp.WithString("ref", mcp.DefaultString("develop"), mcp.Description("Branch name, tag, or commit SHA")),
	)

	commitDetailsTool := mcp.NewTool("gitlab_get_commit_details",
		mcp.WithDescription("Get details of a commit"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or path")),
		mcp.WithString("commit_sha", mcp.Required(), mcp.Description("Commit SHA")),
	)

	s.AddTool(searchTool, util.ErrorGuard(searchProjectsHandler))
	s.AddTool(projectTool, util.ErrorGuard(getProjectHandler))
	s.AddTool(mrListTool, util.ErrorGuard(listMergeRequestsHandler))
	s.AddTool(mrDetailsTool, util.ErrorGuard(getMergeRequestHandler))
	s.AddTool(mrCommentTool, util.ErrorGuard(commentOnMergeRequestHandler))
	s.AddTool(mrChangesTool, util.ErrorGuard(getMergeRequestChangesHandler))
	s.AddTool(fileContentTool, util.ErrorGuard(getFileContentHandler))
	s.AddTool(pipelineTool, util.ErrorGuard(listPipelinesHandler))
	s.AddTool(commitsTool, util.ErrorGuard(listCommitsHandler))
	s.AddTool(commitDetailsTool, util.ErrorGuard(getCommitDetailsHandler))
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
	
	state := "all"
	if value, ok := arguments["state"]; ok {
		state = value.(string)
	}

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

	ref := "develop"
	if value, ok := arguments["ref"]; ok {
		ref = value.(string)
	}

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

func listCommitsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	since, ok := arguments["since"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: since")
	}

	until, ok := arguments["until"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: until")
	}

	ref := "develop"
	if value, ok := arguments["ref"]; ok {
		ref = value.(string)
	}

	sinceTime, err := time.Parse("2006-01-02", since)
	if err != nil {
		return nil, fmt.Errorf("invalid since date: %v", err)
	}

	untilTime, err := time.Parse("2006-01-02 15:04:05", until+" 23:00:00")
	if err != nil {
		return nil, fmt.Errorf("invalid until date: %v", err)
	}

	opt := &gitlab.ListCommitsOptions{
		Since: gitlab.Ptr(sinceTime),
		Until: gitlab.Ptr(untilTime),
		RefName: gitlab.Ptr(ref),
	}

	commits, _, err := gitlabClient().Commits.ListCommits(projectID, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commits for project %s between %s and %s (ref: %s):\n\n", 
		projectID, since, until, ref))

	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ID))
		result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
		result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
		result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
		if commit.LastPipeline != nil {
			result.WriteString("Last Pipeline: \n")
			result.WriteString(fmt.Sprintf("Status: %s\n", commit.LastPipeline.Status))
			result.WriteString(fmt.Sprintf("Ref: %s\n", commit.LastPipeline.Ref))
			result.WriteString(fmt.Sprintf("SHA: %s\n", commit.LastPipeline.SHA))
			result.WriteString(fmt.Sprintf("Created: %s\n", commit.LastPipeline.CreatedAt.Format("2006-01-02 15:04:05")))

		}
	}

	return mcp.NewToolResultText(result.String()), nil
}


func getCommitDetailsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectID := arguments["project_id"].(string)
	commitSHA := arguments["commit_sha"].(string)

	commit, _, err := gitlabClient().Commits.GetCommit(projectID, commitSHA, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %v", err)
	}

	diffs, _, err := gitlabClient().Commits.GetCommitDiff(projectID, commitSHA, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit diffs: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Commit: %s\n", commit.ShortID))
	result.WriteString(fmt.Sprintf("Author: %s\n", commit.AuthorName))
	result.WriteString(fmt.Sprintf("Date: %s\n", commit.CommittedDate.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Message: %s\n", commit.Title))
	result.WriteString(fmt.Sprintf("URL: %s\n\n", commit.WebURL))

	if commit.ParentIDs != nil {
		result.WriteString("Parents:\n")
		for _, parentID := range commit.ParentIDs {
			result.WriteString(fmt.Sprintf("- %s\n", parentID))
		}
		result.WriteString("\n")
	}

	result.WriteString("Diffs:\n")
	for _, diff := range diffs {
		result.WriteString(fmt.Sprintf("File: %s\n", diff.NewPath))
		result.WriteString(fmt.Sprintf("Status: %s\n", getDiffStatus(diff)))

		if diff.Diff != "" {
			result.WriteString("```diff\n")
			result.WriteString(diff.Diff)
			result.WriteString("\n```\n")
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getDiffStatus(diff *gitlab.Diff) string {
	if diff.NewFile {
		return "Added"
	}
	if diff.DeletedFile {
		return "Deleted"
	}
	if diff.RenamedFile {
		return fmt.Sprintf("Renamed from %s", diff.OldPath)
	}
	return "Modified"
}