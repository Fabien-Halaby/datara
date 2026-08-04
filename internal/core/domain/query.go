package domain

import "strings"

// SQLQuery is an immutable value object that represents a SQL query. It can only
// be constructed through NewSQLQuery, which guarantees the query is non-empty and
// trimmed. Whether the query is *safe* to run (read-only, allowed tables, etc.) is
// the responsibility of SQLValidator — this type only guarantees basic structural
// sanity, never emptiness or nil.
type SQLQuery struct {
	raw string
}

// NewSQLQuery builds a SQLQuery from a raw string, rejecting empty input.
func NewSQLQuery(raw string) (SQLQuery, error) {
	cleaned := clean(raw)
	if isEmpty(cleaned) {
		return SQLQuery{}, ErrEmptyQuery
	}

	return SQLQuery{raw: cleaned}, nil
}

// Raw returns the underlying SQL text.
func (q SQLQuery) Raw() string {
	return q.raw
}

func (q SQLQuery) IsZero() bool {
	return q.raw == ""
}

// Helper function:
func clean(raw string) string {
	return strings.TrimSpace(raw)
}

func isEmpty(cleaned string) bool {
	return len(cleaned) == 0
}
