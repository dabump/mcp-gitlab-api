# AGENTS.md

## Repo Shape
- Go MCP server module `github.com/dabump/mcp-gitlab-api`; the executable entrypoint is `cmd/mcp-gitlab-api/main.go`.
- Tool registration is centralized in `internal/tools/common.go` via `RegisterAll`; add new GitLab MCP tools in the relevant `internal/tools/*.go` file and register the group there.
- GitLab HTTP calls go through `internal/gitlabapi.Client.Do`, which returns status, headers, and decoded JSON, or base64 for non-JSON bodies.
- Config loading lives in `internal/config/config.go`; env vars `GITLAB_URL` and `GITLAB_TOKEN` override YAML values.

## Commands
- Use Go `1.26.2` to match `go.mod`.
- Run all verification with `go test ./...`; the Makefile target `make unit-test` also clears test cache and writes `cover.out`.
- Build the CLI with `go build ./cmd/mcp-gitlab-api` or `make build` to create `bin/mcp-gitlab-api`.
- Format with `gofumpt -l -w .`; `make format` installs `gofumpt@latest` before running it.
- Lint with `make golangci-lint`; it installs `golangci-lint@latest` and then runs `golangci-lint run -v`.

## Runtime And Config
- Default run command is `go run ./cmd/mcp-gitlab-api --config config.yaml`.
- For env-only config, use `GITLAB_TOKEN=... go run ./cmd/mcp-gitlab-api --allow-missing-config`; without that flag, the default `config.yaml` path must exist.
- `config.yaml` is gitignored and may contain a token; use `config.example.yaml` as the safe template.
- Default transport is stdio; HTTP mode uses `server.transport: "http"` and serves streamable MCP at `server.endpoint` on `server.host:server.port`.

## Implementation Notes
- `project_id` accepts numeric IDs or path-like IDs; pass it through `gitlabapi.ProjectPath` before embedding it in GitLab API paths.
- Repository file paths and similar path parameters should use `gitlabapi.FilePath` before embedding them in API paths.
- Pagination options are standardized as `page` and `per_page` via `pageOptions()` and `pagination()`.
- Tool annotations matter: use `readOnly(...)` for safe reads and `writeTool(...)` for mutating GitLab actions.
- There are currently no committed `_test.go` files; add focused package tests when changing parsing, config, path escaping, or request construction logic.
