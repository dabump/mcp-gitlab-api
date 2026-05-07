package gitlabapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/dabump/mcp-gitlab-api/internal/config"
)

type Client struct {
	gitlab *gitlab.Client
}

type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Body       any         `json:"body,omitempty"`
}

type Pagination struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	NextPage   int `json:"next_page,omitempty"`
	PrevPage   int `json:"prev_page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	Total      int `json:"total,omitempty"`
}

func New(cfg config.GitLabConfig) (*Client, error) {
	client, err := gitlab.NewClient(cfg.Token, gitlab.WithBaseURL(cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("create GitLab client: %w", err)
	}
	return &Client{gitlab: client}, nil
}

func (c *Client) Do(method, path string, query map[string]any, body any) (Response, error) {
	endpoint := addQuery(path, query)
	req, err := c.gitlab.NewRequest(method, endpoint, body, nil)
	if err != nil {
		return Response{}, fmt.Errorf("create GitLab request: %w", err)
	}

	var responseBody bytes.Buffer
	resp, err := c.gitlab.Do(req, &responseBody)
	if err != nil {
		return Response{}, fmt.Errorf("GitLab request failed: %w", err)
	}
	data := responseBody.Bytes()

	result := Response{Status: resp.Status, StatusCode: resp.StatusCode, Headers: resp.Header, Pagination: paginationFromHeaders(resp.Header)}
	if len(bytes.TrimSpace(data)) == 0 {
		return result, nil
	}

	var decoded any
	if err := json.Unmarshal(data, &decoded); err == nil {
		result.Body = decoded
		return result, nil
	}

	result.Body = map[string]string{
		"encoding": "base64",
		"data":     base64.StdEncoding.EncodeToString(data),
	}
	return result, nil
}

func ProjectPath(projectID string) string {
	return gitlab.PathEscape(projectID)
}

func FilePath(filePath string) string {
	return gitlab.PathEscape(filePath)
}

func addQuery(path string, query map[string]any) string {
	values := url.Values{}
	for key, value := range query {
		addValue(values, key, value)
	}
	if len(values) == 0 {
		return path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + values.Encode()
}

func paginationFromHeaders(headers http.Header) *Pagination {
	pagination := Pagination{
		Page:       headerInt(headers, "X-Page"),
		PerPage:    headerInt(headers, "X-Per-Page"),
		NextPage:   headerInt(headers, "X-Next-Page"),
		PrevPage:   headerInt(headers, "X-Prev-Page"),
		TotalPages: headerInt(headers, "X-Total-Pages"),
		Total:      headerInt(headers, "X-Total"),
	}
	if pagination == (Pagination{}) {
		return nil
	}
	return &pagination
}

func headerInt(headers http.Header, key string) int {
	value := strings.TrimSpace(headers.Get(key))
	if value == "" {
		return 0
	}
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return 0
	}
	return n
}

func addValue(values url.Values, key string, value any) {
	switch v := value.(type) {
	case nil:
		return
	case string:
		if v != "" {
			values.Set(key, v)
		}
	case []string:
		for _, item := range v {
			if item != "" {
				values.Add(key+"[]", item)
			}
		}
	case []any:
		for _, item := range v {
			values.Add(key+"[]", fmt.Sprint(item))
		}
	default:
		values.Set(key, fmt.Sprint(v))
	}
}
