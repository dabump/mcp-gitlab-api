package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
)

func (r Registry) registerIssueTools(s *server.MCPServer) {
	r.add(s, mcp.NewTool("gitlab_list_issues", readOnly(mergeOptions([]mcp.ToolOption{
		mcp.WithDescription("List issues globally or for a project"),
		mcp.WithString("project_id", mcp.Description("Optional project ID/path")),
		mcp.WithString("state", mcp.Description("opened, closed, or all")),
		mcp.WithString("search", mcp.Description("Search title/description")),
		mcp.WithString("labels", mcp.Description("Comma-separated labels")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		q := pagination(a)
		q["state"] = optionalString(a, "state")
		q["search"] = optionalString(a, "search")
		q["labels"] = optionalString(a, "labels")
		if projectID := optionalString(a, "project_id"); projectID != "" {
			return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/issues", gitlabapi.ProjectPath(projectID)), q, nil)
		}
		return r.client.Do(http.MethodGet, "issues", q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_issue", readOnly(mergeOptions(baseProjectOptions("Read issue details"), []mcp.ToolOption{mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.issueGet(a, "")
	})

	r.add(s, mcp.NewTool("gitlab_create_issue", writeTool(mergeOptions(baseProjectOptions("Create an issue"), []mcp.ToolOption{
		mcp.WithString("title", mcp.Required(), mcp.Description("Issue title")),
		mcp.WithString("description", mcp.Description("Issue description")),
		mcp.WithString("labels", mcp.Description("Comma-separated labels")),
		mcp.WithInteger("assignee_id", mcp.Description("Assignee user ID")),
		mcp.WithInteger("milestone_id", mcp.Description("Milestone ID")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/issues", project), nil, bodyFromKeys(a, "title", "description", "labels", "assignee_id", "milestone_id"))
	})

	r.add(s, mcp.NewTool("gitlab_update_issue", writeTool(mergeOptions(baseProjectOptions("Update an issue"), []mcp.ToolOption{
		mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")),
		mcp.WithString("title", mcp.Description("Issue title")),
		mcp.WithString("description", mcp.Description("Issue description")),
		mcp.WithString("state_event", mcp.Description("close or reopen")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := issueIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPut, fmt.Sprintf("projects/%s/issues/%d", project, iid), nil, bodyFromKeys(a, "title", "description", "state_event"))
	})

	r.add(s, mcp.NewTool("gitlab_create_issue_comment", writeTool(mergeOptions(baseProjectOptions("Add a comment to an issue"), []mcp.ToolOption{
		mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := issueIDs(a)
		if err != nil {
			return nil, err
		}
		body, err := requiredString(a, "body")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/issues/%d/notes", project, iid), nil, map[string]any{"body": body})
	})

	r.add(s, mcp.NewTool("gitlab_update_issue_labels", writeTool(mergeOptions(baseProjectOptions("Change issue labels"), []mcp.ToolOption{mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")), mcp.WithString("labels", mcp.Required(), mcp.Description("Comma-separated replacement labels"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.updateIssueFields(a, "labels")
	})

	r.add(s, mcp.NewTool("gitlab_update_issue_milestone", writeTool(mergeOptions(baseProjectOptions("Change issue milestone"), []mcp.ToolOption{mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")), mcp.WithInteger("milestone_id", mcp.Required(), mcp.Description("Milestone ID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.updateIssueFields(a, "milestone_id")
	})

	r.add(s, mcp.NewTool("gitlab_assign_issue", writeTool(mergeOptions(baseProjectOptions("Assign users to an issue"), []mcp.ToolOption{mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")), mcp.WithArray("assignee_ids", mcp.Required(), mcp.Description("Assignee user IDs"), mcp.WithNumberItems())})...)...), func(_ context.Context, a args) (any, error) {
		return r.updateIssueFields(a, "assignee_ids")
	})

	r.add(s, mcp.NewTool("gitlab_link_issue_to_merge_request", readOnly(mergeOptions(baseProjectOptions("List MRs related to/closing an issue"), []mcp.ToolOption{
		mcp.WithInteger("issue_iid", mcp.Required(), mcp.Description("Issue IID")),
		mcp.WithString("relationship", mcp.Description("related or closing")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, iid, err := issueIDs(a)
		if err != nil {
			return nil, err
		}
		rel := optionalString(a, "relationship")
		path := fmt.Sprintf("projects/%s/issues/%d/related_merge_requests", project, iid)
		if rel == "closing" {
			path = fmt.Sprintf("projects/%s/issues/%d/closed_by", project, iid)
		}
		return r.client.Do(http.MethodGet, path, pagination(a), nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_epics", readOnly(mergeOptions([]mcp.ToolOption{
		mcp.WithDescription("Access group epics/roadmaps on GitLab Premium/Ultimate"),
		mcp.WithString("group_id", mcp.Required(), mcp.Description("Group ID or path")),
		mcp.WithString("state", mcp.Description("opened, closed, or all")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		groupID, err := requiredString(a, "group_id")
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["state"] = optionalString(a, "state")
		return r.client.Do(http.MethodGet, fmt.Sprintf("groups/%s/epics", gitlabapi.ProjectPath(groupID)), q, nil)
	})
}

func (r Registry) issueGet(a args, suffix string) (any, error) {
	project, iid, err := issueIDs(a)
	if err != nil {
		return nil, err
	}
	return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/issues/%d%s", project, iid, suffix), pagination(a), nil)
}

func (r Registry) updateIssueFields(a args, fields ...string) (any, error) {
	project, iid, err := issueIDs(a)
	if err != nil {
		return nil, err
	}
	return r.client.Do(http.MethodPut, fmt.Sprintf("projects/%s/issues/%d", project, iid), nil, bodyFromKeys(a, fields...))
}

func issueIDs(a args) (string, int64, error) {
	project, err := projectPath(a)
	if err != nil {
		return "", 0, err
	}
	iid, err := requiredInt(a, "issue_iid")
	if err != nil {
		return "", 0, err
	}
	return project, iid, nil
}
