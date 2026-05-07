package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
)

func (r Registry) registerUserTools(s *server.MCPServer) {
	r.add(s, mcp.NewTool("gitlab_list_project_members", readOnly(mergeOptions(baseProjectOptions("List project members"), []mcp.ToolOption{mcp.WithString("query", mcp.Description("Search members"))}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["query"] = optionalString(a, "query")
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/members/all", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_user", readOnly([]mcp.ToolOption{mcp.WithDescription("Read a user profile"), mcp.WithInteger("user_id", mcp.Required(), mcp.Description("User ID"))}...)...), func(_ context.Context, a args) (any, error) {
		userID, err := requiredInt(a, "user_id")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("users/%d", userID), nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_code_owners", readOnly(mergeOptions(baseProjectOptions("Identify code owners by reading common CODEOWNERS paths"), []mcp.ToolOption{mcp.WithString("ref", mcp.Required(), mcp.Description("Branch, tag, or commit SHA"))})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		ref, err := requiredString(a, "ref")
		if err != nil {
			return nil, err
		}
		paths := []string{"CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"}
		results := make(map[string]any, len(paths))
		for _, filePath := range paths {
			resp, err := r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/files/%s", project, gitlabapi.FilePath(filePath)), map[string]any{"ref": ref}, nil)
			if err != nil {
				results[filePath] = map[string]any{"error": err.Error()}
				continue
			}
			results[filePath] = resp
		}
		return results, nil
	})

	r.add(s, mcp.NewTool("gitlab_resolve_user", readOnly(mergeOptions([]mcp.ToolOption{
		mcp.WithDescription("Resolve usernames/emails to GitLab users"),
		mcp.WithString("search", mcp.Required(), mcp.Description("Username, name, or email search")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		search, err := requiredString(a, "search")
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["search"] = search
		return r.client.Do(http.MethodGet, "users", q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_find_project_maintainers", readOnly(mergeOptions(baseProjectOptions("Find project maintainers"), pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["access_level"] = 40
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/members/all", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_find_reviewers", readOnly(mergeOptions(baseProjectOptions("Find potential reviewers for a project"), []mcp.ToolOption{mcp.WithString("query", mcp.Description("Optional member search query"))}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["query"] = optionalString(a, "query")
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/members/all", project), q, nil)
	})
}
