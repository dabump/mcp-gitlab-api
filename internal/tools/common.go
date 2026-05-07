package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
)

type Registry struct {
	client *gitlabapi.Client
}

type handler func(context.Context, args) (any, error)

type args map[string]any

func RegisterAll(s *server.MCPServer, client *gitlabapi.Client) {
	r := Registry{client: client}
	r.registerRepositoryTools(s)
	r.registerMergeRequestTools(s)
	r.registerIssueTools(s)
	r.registerPipelineTools(s)
	r.registerUserTools(s)
}

func (r Registry) add(s *server.MCPServer, tool mcp.Tool, h handler) {
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := h(ctx, args(req.GetArguments()))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(result)
	})
}

func jsonResult(value any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

func requiredString(a args, name string) (string, error) {
	value, ok := a[name]
	if !ok {
		return "", fmt.Errorf("%s is required", name)
	}
	s, ok := value.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s must be a non-empty string", name)
	}
	return strings.TrimSpace(s), nil
}

func optionalString(a args, name string) string {
	value, ok := a[name]
	if !ok {
		return ""
	}
	s, _ := value.(string)
	return strings.TrimSpace(s)
}

func requiredInt(a args, name string) (int64, error) {
	value, ok := a[name]
	if !ok {
		return 0, fmt.Errorf("%s is required", name)
	}
	return asInt64(value, name)
}

func optionalInt(a args, name string) int64 {
	value, ok := a[name]
	if !ok {
		return 0
	}
	n, _ := asInt64(value, name)
	return n
}

func asInt64(value any, name string) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("%s must be an integer", name)
	}
}

func optionalBool(a args, name string) *bool {
	value, ok := a[name]
	if !ok {
		return nil
	}
	v, ok := value.(bool)
	if !ok {
		return nil
	}
	return &v
}

func pagination(a args) map[string]any {
	q := map[string]any{}
	if page := optionalInt(a, "page"); page > 0 {
		q["page"] = page
	}
	if perPage := optionalInt(a, "per_page"); perPage > 0 {
		q["per_page"] = perPage
	}
	return q
}

func projectPath(a args) (string, error) {
	projectID, err := requiredString(a, "project_id")
	if err != nil {
		return "", err
	}
	return gitlabapi.ProjectPath(projectID), nil
}

func readOnly(opts ...mcp.ToolOption) []mcp.ToolOption {
	return append(opts, mcp.WithReadOnlyHintAnnotation(true), mcp.WithDestructiveHintAnnotation(false))
}

func writeTool(opts ...mcp.ToolOption) []mcp.ToolOption {
	return append(opts, mcp.WithReadOnlyHintAnnotation(false), mcp.WithDestructiveHintAnnotation(true))
}

func baseProjectOptions(description string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithDescription(description),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID or URL-encoded path, for example group/project")),
	}
}

func pageOptions() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithInteger("page", mcp.Description("Page number"), mcp.Min(1)),
		mcp.WithInteger("per_page", mcp.Description("Items per page"), mcp.Min(1), mcp.Max(100)),
	}
}

func mergeOptions(first []mcp.ToolOption, rest ...[]mcp.ToolOption) []mcp.ToolOption {
	out := append([]mcp.ToolOption{}, first...)
	for _, options := range rest {
		out = append(out, options...)
	}
	return out
}

func bodyFromKeys(a args, keys ...string) map[string]any {
	body := map[string]any{}
	for _, key := range keys {
		if value, ok := a[key]; ok {
			body[key] = value
		}
	}
	return body
}
