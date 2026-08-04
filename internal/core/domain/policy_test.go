package domain_test

import (
	"testing"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

func TestSecurityPolicy_IsTableAllowed_EmptyMeansUnrestricted(t *testing.T) {
	policy := domain.SecurityPolicy{}
	if !policy.IsTableAllowed("anything") {
		t.Fatal("expected an empty AllowedTables list to allow any table")
	}
}

func TestSecurityPolicy_IsTableAllowed_RestrictsToList(t *testing.T) {
	policy := domain.SecurityPolicy{AllowedTables: []string{"customers", "products"}}

	if !policy.IsTableAllowed("customers") {
		t.Fatal("expected 'customers' to be allowed")
	}
	if policy.IsTableAllowed("orders") {
		t.Fatal("expected 'orders' to be rejected")
	}
}

func TestDefaultReadOnlyPolicy(t *testing.T) {
	policy := domain.DefaultReadOnlyPolicy()

	if !policy.ReadOnly {
		t.Fatal("expected the default policy to be read-only")
	}
	if policy.MaxRows != 1000 {
		t.Fatalf("expected default MaxRows to be 1000, got %d", policy.MaxRows)
	}
}