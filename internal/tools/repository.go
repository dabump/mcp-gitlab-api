package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
)

func (r Registry) registerRepositoryTools(s *server.MCPServer) {
	r.add(s, mcp.NewTool("gitlab_list_projects", readOnly(mergeOptions(
		[]mcp.ToolOption{mcp.WithDescription("Browse GitLab projects/repositories"), mcp.WithString("search", mcp.Description("Search projects by name/path")), mcp.WithString("membership", mcp.Description("Only projects where the authenticated user is a member: true/false"))},
		pageOptions(),
	)...)...), func(_ context.Context, a args) (any, error) {
		q := pagination(a)
		q["search"] = optionalString(a, "search")
		q["membership"] = optionalString(a, "membership")
		return r.client.Do(http.MethodGet, "projects", q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_project", readOnly(baseProjectOptions("Read project details")...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, "projects/"+project, nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_repository_tree", readOnly(mergeOptions(baseProjectOptions("Read files and directories in a repository tree"), []mcp.ToolOption{
		mcp.WithString("path", mcp.Description("Directory path")),
		mcp.WithString("ref", mcp.Description("Branch, tag, or commit SHA")),
		mcp.WithBoolean("recursive", mcp.Description("Return tree recursively")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["path"] = optionalString(a, "path")
		q["ref"] = optionalString(a, "ref")
		if v := optionalBool(a, "recursive"); v != nil {
			q["recursive"] = *v
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/tree", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_file", readOnly(mergeOptions(baseProjectOptions("Read a repository file"), []mcp.ToolOption{
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path in the repository")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch, tag, or commit SHA")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		filePath, err := requiredString(a, "file_path")
		if err != nil {
			return nil, err
		}
		ref, err := requiredString(a, "ref")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/files/%s", project, gitlabapi.FilePath(filePath)), map[string]any{"ref": ref}, nil)
	})

	r.add(s, mcp.NewTool("gitlab_search_code", readOnly(mergeOptions(
		[]mcp.ToolOption{mcp.WithDescription("Search code across repositories or within a project"), mcp.WithString("search", mcp.Required(), mcp.Description("Search query")), mcp.WithString("project_id", mcp.Description("Optional project ID/path to limit the search")), mcp.WithString("ref", mcp.Description("Optional project ref"))},
		pageOptions(),
	)...)...), func(_ context.Context, a args) (any, error) {
		query, err := requiredString(a, "search")
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["scope"] = "blobs"
		q["search"] = query
		q["ref"] = optionalString(a, "ref")
		if projectID := optionalString(a, "project_id"); projectID != "" {
			return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/search", gitlabapi.ProjectPath(projectID)), q, nil)
		}
		return r.client.Do(http.MethodGet, "search", q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_commits", readOnly(mergeOptions(baseProjectOptions("Get commit history"), []mcp.ToolOption{
		mcp.WithString("ref_name", mcp.Description("Branch, tag, or commit SHA")),
		mcp.WithString("path", mcp.Description("Only commits touching this path")),
		mcp.WithString("since", mcp.Description("ISO8601 start time")),
		mcp.WithString("until", mcp.Description("ISO8601 end time")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["ref_name"] = optionalString(a, "ref_name")
		q["path"] = optionalString(a, "path")
		q["since"] = optionalString(a, "since")
		q["until"] = optionalString(a, "until")
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/commits", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_branches", readOnly(mergeOptions(baseProjectOptions("View repository branches"), pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/branches", project), pagination(a), nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_tags", readOnly(mergeOptions(baseProjectOptions("View repository tags"), pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/tags", project), pagination(a), nil)
	})

	r.add(s, mcp.NewTool("gitlab_compare_refs", readOnly(mergeOptions(baseProjectOptions("Compare commits, branches, or tags"), []mcp.ToolOption{
		mcp.WithString("from", mcp.Required(), mcp.Description("Source ref")),
		mcp.WithString("to", mcp.Required(), mcp.Description("Target ref")),
		mcp.WithBoolean("straight", mcp.Description("Use direct comparison")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		from, err := requiredString(a, "from")
		if err != nil {
			return nil, err
		}
		to, err := requiredString(a, "to")
		if err != nil {
			return nil, err
		}
		q := map[string]any{"from": from, "to": to}
		if v := optionalBool(a, "straight"); v != nil {
			q["straight"] = *v
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/compare", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_commit_diff", readOnly(mergeOptions(baseProjectOptions("Retrieve diffs/patches for a commit"), []mcp.ToolOption{mcp.WithString("sha", mcp.Required(), mcp.Description("Commit SHA"))}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		sha, err := requiredString(a, "sha")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/commits/%s/diff", project, gitlabapi.FilePath(sha)), pagination(a), nil)
	})

	r.add(s, mcp.NewTool("gitlab_blame_file", readOnly(mergeOptions(baseProjectOptions("Blame/annotate file lines"), []mcp.ToolOption{
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path in the repository")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch, tag, or commit SHA")),
		mcp.WithInteger("range_start", mcp.Description("First line to blame")),
		mcp.WithInteger("range_end", mcp.Description("Last line to blame")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		filePath, err := requiredString(a, "file_path")
		if err != nil {
			return nil, err
		}
		ref, err := requiredString(a, "ref")
		if err != nil {
			return nil, err
		}
		q := map[string]any{"ref": ref}
		if n := optionalInt(a, "range_start"); n > 0 {
			q["range[start]"] = n
		}
		if n := optionalInt(a, "range_end"); n > 0 {
			q["range[end]"] = n
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/repository/files/%s/blame", project, gitlabapi.FilePath(filePath)), q, nil)
	})
}
