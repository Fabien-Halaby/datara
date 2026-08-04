package ports

import "github.com/Fabien-Halaby/datara/internal/core/domain"

// SQLValidator inspects a raw SQL string against a SecurityPolicy and
// either returns a validated SQLQuery or an error explaining why the query
// was rejected. Implementations are expected to be dialect-specific
// (Postgres, MySQL, ...) since each SQL dialect has its own grammar.
type SQLValidator interface {
	Validate(raw string) (domain.SQLQuery, error)
}