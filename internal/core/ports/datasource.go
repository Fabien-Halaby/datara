package ports

import (
	"context"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

// DataSource is the boundary between Datara's core and a concrete database
// engine. Each supported engine (Postgres, MySQL, SQLite, ...) implements
// this interface once; nothing else in the codebase needs to change when a
// new engine is added.
type DataSource interface {
	// Ping verifies the connection is alive.
	Ping(ctx context.Context) error

	// Execute runs an already-validated query and returns its result.
	Excute(ctx context.Context, query domain.SQLQuery) (domain.QueryResult, error)

	// Close releases any resources held by the data source.
	Close()
}