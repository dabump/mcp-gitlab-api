package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
)

func (r Registry) registerMergeRequestTools(s *server.MCPServer) {
	r.add(s, mcp.NewTool("gitlab_list_merge_requests", readOnly(mergeOptions([]mcp.ToolOption{
		mcp.WithDescription("List merge requests globally or for a project"),
		mcp.WithString("project_id", mcp.Description("Optional project ID/path")),
		mcp.WithString("state", mcp.Description("opened, closed, locked, merged, or all")),
		mcp.WithString("scope", mcp.Description("created_by_me, assigned_to_me, or all")),
		mcp.WithString("search", mcp.Description("Search title/description")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		q := pagination(a)
		q["state"] = optionalString(a, "state")
		q["scope"] = optionalString(a, "scope")
		q["search"] = optionalString(a, "search")
		if projectID := optionalString(a, "project_id"); projectID != "" {
			return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/merge_requests", gitlabapi.ProjectPath(projectID)), q, nil)
		}
		return r.client.Do(http.MethodGet, "merge_requests", q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_merge_request", readOnly(mergeOptions(baseProjectOptions("Read merge request details"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.mergeRequestGet(a, "")
	})

	r.add(s, mcp.NewTool("gitlab_list_merge_request_discussions", readOnly(mergeOptions(baseProjectOptions("Read merge request discussions/comments"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID"))}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		return r.mergeRequestGet(a, "/discussions")
	})

	r.add(s, mcp.NewTool("gitlab_get_merge_request_changes", readOnly(mergeOptions(baseProjectOptions("Read merge request changes"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.mergeRequestGet(a, "/changes")
	})

	r.add(s, mcp.NewTool("gitlab_create_merge_request", writeTool(mergeOptions(baseProjectOptions("Create a merge request"), []mcp.ToolOption{
		mcp.WithString("source_branch", mcp.Required(), mcp.Description("Source branch")),
		mcp.WithString("target_branch", mcp.Required(), mcp.Description("Target branch")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Merge request title")),
		mcp.WithString("description", mcp.Description("Merge request description")),
		mcp.WithBoolean("remove_source_branch", mcp.Description("Remove source branch when merged")),
		mcp.WithArray("reviewer_ids", mcp.Description("Reviewer user IDs"), mcp.WithNumberItems()),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		body := bodyFromKeys(a, "source_branch", "target_branch", "title", "description", "remove_source_branch", "reviewer_ids")
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/merge_requests", project), nil, body)
	})

	r.add(s, mcp.NewTool("gitlab_create_merge_request_comment", writeTool(mergeOptions(baseProjectOptions("Comment/review on a merge request"), []mcp.ToolOption{
		mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := mergeRequestIDs(a)
		if err != nil {
			return nil, err
		}
		body, err := requiredString(a, "body")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/merge_requests/%d/notes", project, iid), nil, map[string]any{"body": body})
	})

	r.add(s, mcp.NewTool("gitlab_approve_merge_request", writeTool(mergeOptions(baseProjectOptions("Approve a merge request"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID")), mcp.WithString("sha", mcp.Description("Expected HEAD SHA"))})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := mergeRequestIDs(a)
		if err != nil {
			return nil, err
		}
		body := map[string]any{}
		body["sha"] = optionalString(a, "sha")
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/merge_requests/%d/approve", project, iid), nil, body)
	})

	r.add(s, mcp.NewTool("gitlab_unapprove_merge_request", writeTool(mergeOptions(baseProjectOptions("Unapprove a merge request"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID"))})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := mergeRequestIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/merge_requests/%d/unapprove", project, iid), nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_update_merge_request_reviewers", writeTool(mergeOptions(baseProjectOptions("Request reviewers on a merge request"), []mcp.ToolOption{
		mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithArray("reviewer_ids", mcp.Required(), mcp.Description("Reviewer user IDs"), mcp.WithNumberItems()),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := mergeRequestIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPut, fmt.Sprintf("projects/%s/merge_requests/%d", project, iid), nil, bodyFromKeys(a, "reviewer_ids"))
	})

	r.add(s, mcp.NewTool("gitlab_get_merge_request_pipeline", readOnly(mergeOptions(baseProjectOptions("Get CI status/latest pipeline for a merge request"), []mcp.ToolOption{mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.mergeRequestGet(a, "/pipelines")
	})

	r.add(s, mcp.NewTool("gitlab_merge_merge_request", writeTool(mergeOptions(baseProjectOptions("Merge a merge request"), []mcp.ToolOption{
		mcp.WithInteger("merge_request_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithString("merge_commit_message", mcp.Description("Custom merge commit message")),
		mcp.WithBoolean("should_remove_source_branch", mcp.Description("Remove source branch")),
		mcp.WithString("sha", mcp.Description("Expected HEAD SHA")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := mergeRequestIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPut, fmt.Sprintf("projects/%s/merge_requests/%d/merge", project, iid), nil, bodyFromKeys(a, "merge_commit_message", "should_remove_source_branch", "sha"))
	})
}

func (r Registry) mergeRequestGet(a args, suffix string) (any, error) {
	project, iid, err := mergeRequestIDs(a)
	if err != nil {
		return nil, err
	}
	return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/merge_requests/%d%s", project, iid, suffix), pagination(a), nil)
}

func mergeRequestIDs(a args) (string, int64, error) {
	project, err := projectPath(a)
	if err != nil {
		return "", 0, err
	}
	iid, err := requiredInt(a, "merge_request_iid")
	if err != nil {
		return "", 0, err
	}
	return project, iid, nil
}
