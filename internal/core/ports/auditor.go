package ports

import "github.com/Fabien-Halaby/datara/internal/core/domain"

// AuditLogger records what happened to every query Datara has seen,
// whether it was blocked by the security policy or executed against the
// database.
type AuditLogger interface {
	LogBlocked(event domain.QueryBlockedEvent) error
	LogExecuted(event domain.QueryExecutedEvent) error
}