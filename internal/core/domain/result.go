package domain

// QueryResult is the outcome of a successfully executed SQL query.
type QueryResult struct {
	Columns  []string
	Rows     [][]interface{}
	RowCount int

	// Truncated indicates whether the result set was truncated due to a row limit.
	Truncated bool
}
