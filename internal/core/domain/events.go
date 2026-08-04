package domain

import "time"

// QueryBlockedEvent is raised whenever a query fails validation and never
// reaches the database.
type QueryBlockedEvent struct {
	Query     string
	Reason    string
	Timestamp time.Time
}

// QueryExecutedEvent is raised whenever a query is successfully validated
// and executed against the database.
type QueryExecutedEvent struct {
	Query     string
	RowCount  int
	Duration  time.Duration
	Timestamp time.Time
}
