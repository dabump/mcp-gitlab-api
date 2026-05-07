package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (r Registry) registerPipelineTools(s *server.MCPServer) {
	r.add(s, mcp.NewTool("gitlab_list_pipelines", readOnly(mergeOptions(baseProjectOptions("List project pipelines"), []mcp.ToolOption{
		mcp.WithString("ref", mcp.Description("Branch or tag")),
		mcp.WithString("status", mcp.Description("Pipeline status")),
	}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["ref"] = optionalString(a, "ref")
		q["status"] = optionalString(a, "status")
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/pipelines", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_pipeline", readOnly(mergeOptions(baseProjectOptions("Read pipeline status"), []mcp.ToolOption{mcp.WithInteger("pipeline_id", mcp.Required(), mcp.Description("Pipeline ID"))})...)...), func(_ context.Context, a args) (any, error) {
		project, pipeline, err := pipelineIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/pipelines/%d", project, pipeline), nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_job_log", readOnly(mergeOptions(baseProjectOptions("Read CI job logs"), []mcp.ToolOption{mcp.WithInteger("job_id", mcp.Required(), mcp.Description("Job ID"))})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		jobID, err := requiredInt(a, "job_id")
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/jobs/%d/trace", project, jobID), nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_retry_job", writeTool(mergeOptions(baseProjectOptions("Retry a CI job"), []mcp.ToolOption{mcp.WithInteger("job_id", mcp.Required(), mcp.Description("Job ID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.jobAction(a, "retry")
	})

	r.add(s, mcp.NewTool("gitlab_cancel_job", writeTool(mergeOptions(baseProjectOptions("Cancel a CI job"), []mcp.ToolOption{mcp.WithInteger("job_id", mcp.Required(), mcp.Description("Job ID"))})...)...), func(_ context.Context, a args) (any, error) {
		return r.jobAction(a, "cancel")
	})

	r.add(s, mcp.NewTool("gitlab_trigger_pipeline", writeTool(mergeOptions(baseProjectOptions("Trigger a pipeline"), []mcp.ToolOption{
		mcp.WithString("ref", mcp.Required(), mcp.Description("Branch or tag")),
		mcp.WithString("token", mcp.Description("Pipeline trigger token. If omitted, creates a pipeline using the configured PAT.")),
		mcp.WithObject("variables", mcp.Description("Pipeline variables")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		ref, err := requiredString(a, "ref")
		if err != nil {
			return nil, err
		}
		body := bodyFromKeys(a, "variables")
		body["ref"] = ref
		if token := optionalString(a, "token"); token != "" {
			body["token"] = token
			return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/trigger/pipeline", project), nil, body)
		}
		return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/pipeline", project), nil, body)
	})

	r.add(s, mcp.NewTool("gitlab_get_job_artifact", readOnly(mergeOptions(baseProjectOptions("Read/download a job artifact"), []mcp.ToolOption{
		mcp.WithInteger("job_id", mcp.Required(), mcp.Description("Job ID")),
		mcp.WithString("artifact_path", mcp.Description("Optional path inside artifact archive")),
	})...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		jobID, err := requiredInt(a, "job_id")
		if err != nil {
			return nil, err
		}
		path := fmt.Sprintf("projects/%s/jobs/%d/artifacts", project, jobID)
		if artifactPath := optionalString(a, "artifact_path"); artifactPath != "" {
			path = fmt.Sprintf("projects/%s/jobs/%d/artifacts/%s", project, jobID, artifactPath)
		}
		return r.client.Do(http.MethodGet, path, nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_get_pipeline_test_report", readOnly(mergeOptions(baseProjectOptions("Download/read a pipeline test report"), []mcp.ToolOption{mcp.WithInteger("pipeline_id", mcp.Required(), mcp.Description("Pipeline ID"))})...)...), func(_ context.Context, a args) (any, error) {
		project, pipeline, err := pipelineIDs(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/pipelines/%d/test_report", project, pipeline), nil, nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_deployments", readOnly(mergeOptions(baseProjectOptions("Surface deployment info"), []mcp.ToolOption{mcp.WithString("environment", mcp.Description("Environment name")), mcp.WithString("status", mcp.Description("Deployment status"))}, pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		q := pagination(a)
		q["environment"] = optionalString(a, "environment")
		q["status"] = optionalString(a, "status")
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/deployments", project), q, nil)
	})

	r.add(s, mcp.NewTool("gitlab_list_environments", readOnly(mergeOptions(baseProjectOptions("List project environments"), pageOptions())...)...), func(_ context.Context, a args) (any, error) {
		project, err := projectPath(a)
		if err != nil {
			return nil, err
		}
		return r.client.Do(http.MethodGet, fmt.Sprintf("projects/%s/environments", project), pagination(a), nil)
	})
}

func (r Registry) jobAction(a args, action string) (any, error) {
	project, err := projectPath(a)
	if err != nil {
		return nil, err
	}
	jobID, err := requiredInt(a, "job_id")
	if err != nil {
		return nil, err
	}
	return r.client.Do(http.MethodPost, fmt.Sprintf("projects/%s/jobs/%d/%s", project, jobID, action), nil, nil)
}

func pipelineIDs(a args) (string, int64, error) {
	project, err := projectPath(a)
	if err != nil {
		return "", 0, err
	}
	pipelineID, err := requiredInt(a, "pipeline_id")
	if err != nil {
		return "", 0, err
	}
	return project, pipelineID, nil
}
