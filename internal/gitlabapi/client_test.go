package gitlabapi

import (
	"net/http"
	"testing"
)

func TestAddQueryIncludesPaginationValues(t *testing.T) {
	endpoint := addQuery("projects", map[string]any{"page": int64(2), "per_page": 50})

	if endpoint != "projects?page=2&per_page=50" {
		t.Fatalf("endpoint = %q, want %q", endpoint, "projects?page=2&per_page=50")
	}
}

func TestAddQueryAppendsPaginationToExistingQuery(t *testing.T) {
	endpoint := addQuery("projects?membership=true", map[string]any{"page": 3})

	if endpoint != "projects?membership=true&page=3" {
		t.Fatalf("endpoint = %q, want %q", endpoint, "projects?membership=true&page=3")
	}
}

func TestPaginationFromHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Page", "2")
	headers.Set("X-Per-Page", "50")
	headers.Set("X-Next-Page", "3")
	headers.Set("X-Prev-Page", "1")
	headers.Set("X-Total-Pages", "4")
	headers.Set("X-Total", "175")

	pagination := paginationFromHeaders(headers)
	if pagination == nil {
		t.Fatal("pagination = nil, want pagination metadata")
	}
	if pagination.Page != 2 || pagination.PerPage != 50 || pagination.NextPage != 3 || pagination.PrevPage != 1 || pagination.TotalPages != 4 || pagination.Total != 175 {
		t.Fatalf("pagination = %+v, want page/per-page/next/prev/total-pages/total metadata", pagination)
	}
}

func TestPaginationFromHeadersReturnsNilWithoutPaginationHeaders(t *testing.T) {
	if pagination := paginationFromHeaders(http.Header{}); pagination != nil {
		t.Fatalf("pagination = %+v, want nil", pagination)
	}
}

func TestSplitEndpointQuery(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantPath string
		wantQuery string
	}{
		{name: "without query", endpoint: "projects", wantPath: "projects", wantQuery: ""},
		{name: "with query", endpoint: "projects?page=1&per_page=100", wantPath: "projects", wantQuery: "page=1&per_page=100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotQuery := splitEndpointQuery(tt.endpoint)
			if gotPath != tt.wantPath || gotQuery != tt.wantQuery {
				t.Fatalf("splitEndpointQuery(%q) = (%q, %q), want (%q, %q)", tt.endpoint, gotPath, gotQuery, tt.wantPath, tt.wantQuery)
			}
		})
	}
}
