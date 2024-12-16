# MyMCP Server

A powerful Model Context Protocol (MCP) server implementation with integrations for GitLab, Jira, Confluence, YouTube, and more. This server provides AI-powered search capabilities and various utility tools for development workflows.

[Tutorial](https://www.youtube.com/watch?v=XnDFtYKU6xU)

## Community

For community support, discussions, and updates, please visit our forum at [community.aiocean.io](https://community.aiocean.io/).


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
        "QDRANT_HOST": "",
        "ATLASSIAN_HOST": "",
        "ATLASSIAN_EMAIL": "",
        "GITLAB_HOST": "",
        "GITLAB_TOKEN": "",
        "BRAVE_API_KEY": "",
        "ENABLE_TOOLS": "Check environment variable first for backward compatibility",
        "ATLASSIAN_TOKEN": "",
        "GOOGLE_AI_API_KEY": "",
        "PROXY_URL": "",
        "OPENAI_API_KEY": "",
        "QDRANT_PORT": "",
        "GOOGLE_TOKEN_FILE": "",
        "GOOGLE_CREDENTIALS_FILE": "",
        "QDRANT_API_KEY": ""
      }
    }
  }
}
```

## Enable Tools

There are a hidden variable `ENABLE_TOOLS` in the environment variable. It is a comma separated list of tools group to enable. If not set, all tools will be enabled. Leave it empty to enable all tools.


Here is the list of tools group:

- `gemini`: Gemini-powered search
- `fetch`: Fetch tools
- `confluence`: Confluence tools
- `youtube`: YouTube tools
- `jira`: Jira tools
- `gitlab`: GitLab tools
- `script`: Script tools
- `rag`: RAG tools

## Available Tools

### calendar_create_event

Create a new event in Google Calendar

Arguments:

- `summary` (String) (Required): Title of the event
- `description` (String): Description of the event
- `start_time` (String) (Required): Start time of the event in RFC3339 format (e.g., 2023-12-25T09:00:00Z)
- `end_time` (String) (Required): End time of the event in RFC3339 format
- `attendees` (String): Comma-separated list of attendee email addresses

### calendar_list_events

List upcoming events in Google Calendar

Arguments:

- `time_min` (String): Start time for the search in RFC3339 format (default: now)
- `time_max` (String): End time for the search in RFC3339 format (default: 1 week from now)
- `max_results` (Number): Maximum number of events to return (default: 10)

### calendar_update_event

Update an existing event in Google Calendar

Arguments:

- `event_id` (String) (Required): ID of the event to update
- `summary` (String): New title of the event
- `description` (String): New description of the event
- `start_time` (String): New start time of the event in RFC3339 format
- `end_time` (String): New end time of the event in RFC3339 format
- `attendees` (String): Comma-separated list of new attendee email addresses

### calendar_respond_to_event

Respond to an event invitation in Google Calendar

Arguments:

- `event_id` (String) (Required): ID of the event to respond to
- `response` (String) (Required): Your response (accepted, declined, or tentative)

### confluence_search

Search Confluence

Arguments:

- `query` (String) (Required): Atlassian Confluence Query Language (CQL)

### confluence_get_page

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

### gitlab_create_mr

Create a new merge request

Arguments:

- `project_id` (String) (Required): Project ID or path
- `source_branch` (String) (Required): Source branch name
- `target_branch` (String) (Required): Target branch name
- `title` (String) (Required): Merge request title
- `description` (String): Merge request description

### gmail_search

Search emails in Gmail using Gmail's search syntax

Arguments:

- `query` (String) (Required): Gmail search query. Follow Gmail's search syntax

### gmail_move_to_spam

Move specific emails to spam folder in Gmail by message IDs

Arguments:

- `message_ids` (String) (Required): Comma-separated list of message IDs to move to spam

### gmail_create_filter

Create a Gmail filter with specified criteria and actions

Arguments:

- `from` (String): Filter emails from this sender
- `to` (String): Filter emails to this recipient
- `subject` (String): Filter emails with this subject
- `query` (String): Additional search query criteria
- `add_label` (Boolean): Add label to matching messages
- `label_name` (String): Name of the label to add (required if add_label is true)
- `mark_important` (Boolean): Mark matching messages as important
- `mark_read` (Boolean): Mark matching messages as read
- `archive` (Boolean): Archive matching messages

### gmail_list_filters

List all Gmail filters in the account

### gmail_list_labels

List all Gmail labels in the account

### gmail_delete_filter

Delete a Gmail filter by its ID

Arguments:

- `filter_id` (String) (Required): The ID of the filter to delete

### gmail_delete_label

Delete a Gmail label by its ID

Arguments:

- `label_id` (String) (Required): The ID of the label to delete

### jira_get_issue

Get Jira issue details

Arguments:

- `issue_key` (String) (Required): Jira issue key (e.g., KP-2)

### jira_search_issue

Search/list for Jira issues by JQL

Arguments:

- `jql` (String) (Required): JQL query to search/list for Jira issues

### jira_list_sprints

List all sprints in a Jira project

Arguments:

- `board_id` (String) (Required): Jira board ID

### jira_create_issue

Create a new Jira issue

Arguments:

- `project_key` (String) (Required): Jira project key (e.g., KP)
- `summary` (String) (Required): Summary of the issue
- `description` (String) (Required): Description of the issue
- `issue_type` (String) (Required): Type of the issue (e.g., Bug, Task)

### jira_update_issue

Update an existing Jira issue

Arguments:

- `issue_key` (String) (Required): Jira issue key (e.g., KP-2)
- `summary` (String): New summary of the issue
- `description` (String): New description of the issue
- `status` (String): New status of the issue

### RAG_memory_index_content

Index a note into memory, can be inserted or updated

Arguments:

- `collection` (String) (Required): Memory collection name
- `filePath` (String) (Required): note file path
- `payload` (String) (Required): Plain text payload

### RAG_memory_index_file

Index a local file into memory

Arguments:

- `collection` (String) (Required): Memory collection name
- `filePath` (String) (Required): Path to the local file to be indexed

### RAG_memory_create_collection

Create a new vector collection in memory

Arguments:

- `collection` (String) (Required): Memory collection name

### RAG_memory_delete_collection

Delete a vector collection in memory

Arguments:

- `collection` (String) (Required): Memory collection name

### RAG_memory_list_collections

List all vector collections in memory

### RAG_memory_search

Search for notes in a collection based on a query

Arguments:

- `collection` (String) (Required): Memory collection name
- `query` (String) (Required): Search term, should be a keyword or a phrase

### RAG_memory_delete_index_by_filepath

Delete a vector index by filePath

Arguments:

- `collection` (String) (Required): Memory collection name
- `filePath` (String) (Required): Path to the local file to be deleted

### execute_comand_line_script

Execute a script file on user machine. Non interactive. Do not do unsafe operations

Arguments:

- `content` (String) (Required): 
- `interpreter` (String) (Default: /bin/sh): Script interpreter (e.g., /bin/sh, /bin/bash, python, etc.)
- `working_dir` (String): Working directory for script execution

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

