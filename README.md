# MyMCP Server

A powerful Model Context Protocol (MCP) server implementation with integrations for GitLab, Jira, Confluence, YouTube, and more. This server provides AI-powered search capabilities and various utility tools for development workflows.

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
        "ATLASSIAN_EMAIL": "",
        "ATLASSIAN_TOKEN": "",
        "PROXY_URL": "",
        "GOOGLE_AI_API_KEY": "",
        "GITLAB_TOKEN": "",
        "GITLAB_HOST": "",
        "BRAVE_API_KEY": "",
        "ATLASSIAN_HOST": ""
      }
    }
  }
}
```

## Available Tools

### execute_comand_line_script

Execute a script file on user machine. Non interactive. Do not do unsafe operations

Arguments:

- `content` (String) (Required): 
- `interpreter` (String) (Default: /bin/sh): Script interpreter (e.g., /bin/sh, /bin/bash, python, etc.)
- `working_dir` (String): Working directory for script execution

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

### gitlab_list_projects

List GitLab projects

Arguments:

- `group_id` (String) (Required): gitlab group ID
- `search` (String): Multiple terms can be provided, separated by an escaped space, either + or %20, and will be ANDed together. Example: one+two will match substrings one and two (in any order).

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

### gitlab_list_user_events

List GitLab user events within a date range

Arguments:

- `username` (String) (Required): GitLab username
- `since` (String) (Required): Start date (YYYY-MM-DD)
- `until` (String) (Required): End date (YYYY-MM-DD)

### gitlab_list_group_users

List all users in a GitLab group

Arguments:

- `group_id` (String) (Required): GitLab group ID

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

