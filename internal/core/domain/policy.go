package domain

// SecurityPolicy describes the rules a query must satisfy before Datara
// will execute it. This is the real business logic of Datara: everything
// else in the codebase exists to enforce this policy as cheaply and
// reliably as possible.
type SecurityPolicy struct {
	// ReadOnly, when true, only allows single SELECT statements with no
	// data-modifying side effects (no SELECT ... INTO, no locking clauses).
	ReadOnly bool

	// AllowedTables restricts which tables may be referenced. An empty
	// slice means "no restriction" — all tables are allowed.
	AllowedTables []string

	// MaxRows caps the number of rows a single query may return. Zero
	// means "no limit" (not recommended in production).
	MaxRows int
}

// DefaultReadOnlyPolicy returns the strict default policy: read-only,
// unrestricted tables, capped at 1000 rows.
func DefaultReadOnlyPolicy() SecurityPolicy {
	return SecurityPolicy{
		ReadOnly:      true,
		AllowedTables: []string{},
		MaxRows:       1000,
	}
}

// IsTableAllowed reports whether the given table name is permitted under
// this policy.
func (p SecurityPolicy) IsTableAllowed(table string) bool {
	if len(p.AllowedTables) == 0 {
		return true
	}

	for _, allowed := range p.AllowedTables {
		if allowed == table {
			return true
		}
	}

	return false
}
