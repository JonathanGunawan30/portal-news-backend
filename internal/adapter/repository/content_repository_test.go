package repository

import (
	"reflect"
	"testing"
)

func TestSanitizeOrderClause(t *testing.T) {
	tests := []struct {
		orderBy   string
		orderType string
		want      string
	}{
		{"created_at", "DESC", "created_at DESC"},
		{"title", "asc", "title ASC"},
		{"id", "DESC", "id DESC"},
		{"created_at; DROP TABLE contents--", "DESC", "created_at DESC"},
		{"title", "DESC; SELECT 1", "title DESC"},
		{"unknown_col", "ASC", "created_at ASC"},
	}

	for _, tc := range tests {
		got := sanitizeOrderClause(tc.orderBy, tc.orderType)
		if got != tc.want {
			t.Errorf("sanitizeOrderClause(%q, %q) = %q, want %q", tc.orderBy, tc.orderType, got, tc.want)
		}
	}
}

func TestSplitTags(t *testing.T) {
	if got := splitTags(""); !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("expected empty slice for empty tags, got %#v", got)
	}
	if got := splitTags("go, api , news, "); !reflect.DeepEqual(got, []string{"go", "api", "news"}) {
		t.Fatalf("unexpected tags split: %#v", got)
	}
}
