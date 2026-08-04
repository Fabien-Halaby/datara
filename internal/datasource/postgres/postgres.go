package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

// DataSource is the Postgres implementation of ports.DataSource.
type DataSource struct {
	pool    *pgxpool.Pool
	maxRows int
}

// New connects to Postgres using dsn and returns a ready-to-use
// DataSource. maxRows caps how many rows Execute will return, regardless
// of what the query itself requests (0 = no cap).
func New(ctx context.Context, dsn string, maxRows int) (*DataSource, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return &DataSource{pool: pool, maxRows: maxRows}, nil
}

func (d *DataSource) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func (d *DataSource) Execute(ctx context.Context, query domain.SQLQuery) (domain.QueryResult, error) {
	rows, err := d.pool.Query(ctx, query.Raw())
	if err != nil {
		return domain.QueryResult{}, fmt.Errorf("executing query: %w", err)
	}

	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = string(f.Name)
	}

	var resultRows [][]any
	truncated := false
	for rows.Next() {
		if d.maxRows > 0 && len(resultRows) >= d.maxRows {
			truncated = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			return domain.QueryResult{}, fmt.Errorf("reading row: %w", err)
		}
		resultRows = append(resultRows, values)
	}
	if err := rows.Err(); err != nil {
		return domain.QueryResult{}, fmt.Errorf("iterating rows: %w", err)
	}

	return domain.QueryResult{
		Columns:   columns,
		Rows:      resultRows,
		RowCount:  len(resultRows),
		Truncated: truncated,
	}, nil
}

func (d *DataSource) Close() {
	d.pool.Close()
}