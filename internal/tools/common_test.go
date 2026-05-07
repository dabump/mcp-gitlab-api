package tools

import (
	"encoding/json"
	"testing"
)

func TestPaginationAcceptsMCPNumericArgumentTypes(t *testing.T) {
	query := pagination(args{"page": float64(2), "per_page": json.Number("50")})

	if query["page"] != int64(2) {
		t.Fatalf("page = %#v, want int64(2)", query["page"])
	}
	if query["per_page"] != int64(50) {
		t.Fatalf("per_page = %#v, want int64(50)", query["per_page"])
	}
}

func TestPaginationOmitsUnsetValues(t *testing.T) {
	query := pagination(args{})

	if len(query) != 0 {
		t.Fatalf("query = %#v, want empty map", query)
	}
}

func TestPaginationOmitsInvalidValues(t *testing.T) {
	query := pagination(args{"page": "2", "per_page": -1})

	if len(query) != 0 {
		t.Fatalf("query = %#v, want empty map", query)
	}
}
