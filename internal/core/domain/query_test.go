package domain_test

import (
	"testing"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

func TestNewSQLQuery_TrimsWhitespace(t *testing.T) {
	q, err := domain.NewSQLQuery("  SELECT 1  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Raw() != "SELECT 1" {
		t.Fatalf("expected trimmed query, got %q", q.Raw())
	}
}

func TestNewSQLQuery_RejectsEmpty(t *testing.T) {
	cases := []string{"", "   ", "\n\t"}
	for _, c := range cases {
		if _, err := domain.NewSQLQuery(c); err == nil {
			t.Fatalf("expected an error for input %q, got none", c)
		}
	}
}