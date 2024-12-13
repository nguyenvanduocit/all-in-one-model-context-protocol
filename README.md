# MyMCP Server

A powerful Model Context Protocol (MCP) server implementation with integrations for GitLab, Jira, Confluence, YouTube, and more. This server provides AI-powered search capabilities and various utility tools for development workflows.

## Features

- 🔍 AI-powered web search using Google Gemini
- 🎥 YouTube transcript extraction
- 📝 Confluence and Jira integration
- 🦊 GitLab project management and code analysis
- 🌐 Web content fetching
- 💻 Local CLI command execution
- 🔎 Brave Search integration

## Prerequisites

- Go 1.23.2 or higher
- Various API keys and tokens for the services you want to use

## Installation

1. Install the server:
```bash
go install github.com/nguyenvanduocit/all-in-one-model-context-protocol@latest
```

2. Config your claude's config:

```json{claude_desktop_config.json}
{
  "mcpServers": {
    "my_mcp_server": {
      "command": "all-in-one-model-context-protocol",
      "args": [],
      "env": {
        "ATLASSIAN_TOKEN": "Token is the API token of the user",
        "PROXY_URL": "",
        "GOOGLE_AI_API_KEY": "",
        "GITLAB_TOKEN": "",
        "GITLAB_HOST": "",
        "BRAVE_API_KEY": "",
        "ATLASSIAN_HOST": "Host is the URL of the Jira instance",
        "ATLASSIAN_EMAIL": "Mail is the email of the user"
      }
    }
  }
}
```

## Available Tools

### cli_execute

Arguments:

- `command` (String) (Required): Command to execute
- `args` (String): Command arguments (space-separated)
- `working_dir` (String): Working directory for command execution

### confluence_search

Search Confluence

Arguments:

- `query` (String) (Required): Atlassian Confluence Query Language (CQL)

### get_confluence_page

Get Confluence page content

Arguments:

- `page_id` (String) (Required): Confluence page ID

### fetch_url

Fetch/read a http URL and return the content

Arguments:

- `url` (String) (Required): URL to fetch

### ai_web_search

search the web by using Google AI Search. Best tool to update realtime information

Arguments:

- `question` (String) (Required): The question to ask. Should be a question
- `context` (String) (Required): Context/purpose of the question, helps Gemini to understand the question better

### gitlab_search_projects

Search GitLab projects

Arguments:

- `query` (String) (Required): Search query

### gitlab_get_project

Get GitLab project details

Arguments:

- `project_id` (String) (Required): Project ID or path

### gitlab_list_mrs

List merge requests

Arguments:

- `project_id` (String) (Required): Project ID or path
- `state` (String) (Default: all): MR state (opened/closed/merged)

### gitlab_get_mr_details

Get merge request details

Arguments:

- `project_id` (String) (Required): Project ID or path
- `mr_iid` (String) (Required): Merge request IID

### gitlab_create_MR_note

Create a note on a merge request

Arguments:

- `project_id` (String) (Required): Project ID or path
- `mr_iid` (String) (Required): Merge request IID
- `comment` (String) (Required): Comment text

### gitlab_list_mr_changes

Get detailed changes of a merge request

Arguments:

- `project_id` (String) (Required): Project ID or path
- `mr_iid` (String) (Required): Merge request IID

### gitlab_get_file_content

Get file content from a GitLab repository

Arguments:

- `project_id` (String) (Required): Project ID or path
- `file_path` (String) (Required): Path to the file in the repository
- `ref` (String) (Default: develop): Branch name, tag, or commit SHA

### gitlab_list_pipelines

List pipelines for a GitLab project

Arguments:

- `project_id` (String) (Required): Project ID or path
- `status` (String) (Default: all): Pipeline status (running/pending/success/failed/canceled/skipped/all)

### gitlab_list_commits

List commits in a GitLab project within a date range

Arguments:

- `project_id` (String) (Required): Project ID or path
- `since` (String) (Required): Start date (YYYY-MM-DD)
- `until` (String) (Required): End date (YYYY-MM-DD)
- `ref` (String) (Default: develop): Branch name, tag, or commit SHA

### gitlab_get_commit_details

Get details of a commit

Arguments:

- `project_id` (String) (Required): Project ID or path
- `commit_sha` (String) (Required): Commit SHA

### get_jira_issue

Get Jira issue details

Arguments:

- `issue_key` (String) (Required): Jira issue key (e.g., KP-2)

### web_search

Search the web using Brave Search API

Arguments:

- `query` (String) (Required): Query to search for (max 400 chars, 50 words)
- `count` (Number) (Default: 5): Number of results (1-20, default 5)
- `country` (String) (Default: ALL): Country code

### youtube_transcript

Get YouTube video transcript

Arguments:

- `url` (String) (Required): YouTube video URL
- `lang` (String) (Default: en): Language code (default: en)
- `country` (String) (Default: US): Country code (default: US)

