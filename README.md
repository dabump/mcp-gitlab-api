# MCP GitLab API

MCP server for GitLab, written in Go. It uses `gitlab.com/gitlab-org/api/client-go` for authenticated GitLab API access and `github.com/mark3labs/mcp-go` for MCP stdio transport.

## Configuration

Create `config.yaml`:

```yaml
server:
  # stdio is the default for local MCP clients. Use http to bind to host:port.
  transport: "stdio"
  host: "127.0.0.1"
  port: 8080
  endpoint: "/mcp"

gitlab:
  url: "https://gitlab.com"
  token: "glpat-your-token"
```

Environment variables override the file:

- `GITLAB_URL`
- `GITLAB_TOKEN`

## Run

```sh
go run ./cmd/mcp-gitlab-api --config config.yaml
```

For env-only configuration:

```sh
GITLAB_TOKEN=glpat-your-token go run ./cmd/mcp-gitlab-api --allow-missing-config
```

To run the MCP server over streamable HTTP instead of stdio, set:

```yaml
server:
  transport: "http"
  host: "127.0.0.1"
  port: 8080
  endpoint: "/mcp"
```

The MCP endpoint will be `http://127.0.0.1:8080/mcp`.

## Claude Desktop Example

```json
{
  "mcpServers": {
    "gitlab": {
      "command": "/path/to/mcp-gitlab-api",
      "args": ["--config", "/path/to/config.yaml"]
    }
  }
}
```

## OpenCode Example

```json
{
  "mcp": {
    "gitlab": {
      "enabled": true,
      "type": "remote",
      "url": "http://127.0.0.1:8082/mcp"
    }
  }
}
```

## Tools

Repository and code access:

- `gitlab_list_projects`
- `gitlab_get_project`
- `gitlab_list_repository_tree`
- `gitlab_get_file`
- `gitlab_search_code`
- `gitlab_list_commits`
- `gitlab_list_branches`
- `gitlab_list_tags`
- `gitlab_compare_refs`
- `gitlab_get_commit_diff`
- `gitlab_blame_file`
- `gitlab_get_merge_request_changes`

Merge request operations:

- `gitlab_list_merge_requests`
- `gitlab_get_merge_request`
- `gitlab_list_merge_request_discussions`
- `gitlab_create_merge_request`
- `gitlab_create_merge_request_comment`
- `gitlab_approve_merge_request`
- `gitlab_unapprove_merge_request`
- `gitlab_update_merge_request_reviewers`
- `gitlab_get_merge_request_pipeline`
- `gitlab_merge_merge_request`

Issues and project management:

- `gitlab_list_issues`
- `gitlab_get_issue`
- `gitlab_create_issue`
- `gitlab_update_issue`
- `gitlab_create_issue_comment`
- `gitlab_update_issue_labels`
- `gitlab_update_issue_milestone`
- `gitlab_assign_issue`
- `gitlab_link_issue_to_merge_request`
- `gitlab_list_epics`

CI/CD and pipelines:

- `gitlab_list_pipelines`
- `gitlab_get_pipeline`
- `gitlab_get_job_log`
- `gitlab_retry_job`
- `gitlab_cancel_job`
- `gitlab_trigger_pipeline`
- `gitlab_get_job_artifact`
- `gitlab_get_pipeline_test_report`
- `gitlab_list_deployments`
- `gitlab_list_environments`

User and team information:

- `gitlab_list_project_members`
- `gitlab_get_user`
- `gitlab_get_code_owners`
- `gitlab_resolve_user`
- `gitlab_find_project_maintainers`
- `gitlab_find_reviewers`

## Token Scopes

Use the least privilege token that supports your intended operations. Read-only tools generally need `read_api` or `api`; write tools such as creating MRs, commenting, approvals, merges, retries, cancellations, and pipeline triggers need `api`.

Premium-only APIs such as epics return the GitLab API error from your instance if the feature is unavailable or the token lacks access.
