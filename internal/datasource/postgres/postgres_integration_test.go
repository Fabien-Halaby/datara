//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
	"github.com/Fabien-Halaby/datara/internal/datasource/postgres"
)

// startTestPostgres spins up a throwaway Postgres container for the
// duration of a single test and returns a ready-to-use DSN. The container
// is automatically terminated when the test finishes.
//
// Postgres restarts itself once internally right after first-time
// initialization, so we wait for the "ready to accept connections" log
// line to appear twice — otherwise the container can be reported ready
// while it's mid-restart, causing "connection reset by peer" errors.
func startTestPostgres(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("datara_test"),
		tcpostgres.WithUsername("datara"),
		tcpostgres.WithPassword("datara"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	return dsn
}

// seedSchema creates a small customers table with a known, fixed dataset
// so assertions below can rely on exact values.
func seedSchema(t *testing.T, dsn string) {
	t.Helper()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect for seeding: %v", err)
	}
	defer pool.Close()

	ddl := `
		CREATE TABLE customers (
			id    SERIAL PRIMARY KEY,
			name  TEXT NOT NULL,
			email TEXT NOT NULL
		);
		INSERT INTO customers (name, email) VALUES
			('Alice', 'alice@example.com'),
			('Bob', 'bob@example.com'),
			('Carol', 'carol@example.com');
	`
	if _, err := pool.Exec(ctx, ddl); err != nil {
		t.Fatalf("failed to seed schema: %v", err)
	}
}

func mustQuery(t *testing.T, sql string) domain.SQLQuery {
	t.Helper()
	q, err := domain.NewSQLQuery(sql)
	if err != nil {
		t.Fatalf("failed to build SQLQuery: %v", err)
	}
	return q
}

func TestPostgresDataSource_PingSucceeds(t *testing.T) {
	dsn := startTestPostgres(t)

	ds, err := postgres.New(context.Background(), dsn, 0)
	if err != nil {
		t.Fatalf("failed to connect DataSource: %v", err)
	}
	defer ds.Close()

	if err := ds.Ping(context.Background()); err != nil {
		t.Fatalf("expected Ping to succeed, got: %v", err)
	}
}

func TestPostgresDataSource_ExecuteReturnsRealRows(t *testing.T) {
	dsn := startTestPostgres(t)
	seedSchema(t, dsn)

	ds, err := postgres.New(context.Background(), dsn, 0)
	if err != nil {
		t.Fatalf("failed to connect DataSource: %v", err)
	}
	defer ds.Close()

	query := mustQuery(t, "SELECT id, name, email FROM customers ORDER BY id")

	result, err := ds.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error executing query: %v", err)
	}

	if result.RowCount != 3 {
		t.Fatalf("expected 3 rows, got %d", result.RowCount)
	}
	if result.Truncated {
		t.Fatal("did not expect the result to be truncated")
	}

	wantColumns := []string{"id", "name", "email"}
	if len(result.Columns) != len(wantColumns) {
		t.Fatalf("expected %d columns, got %d", len(wantColumns), len(result.Columns))
	}
	for i, col := range wantColumns {
		if result.Columns[i] != col {
			t.Errorf("expected column %d to be %q, got %q", i, col, result.Columns[i])
		}
	}

	firstRowName, ok := result.Rows[0][1].(string)
	if !ok || firstRowName != "Alice" {
		t.Errorf("expected first row's name to be %q, got %v", "Alice", result.Rows[0][1])
	}
}

func TestPostgresDataSource_ExecuteRespectsMaxRows(t *testing.T) {
	dsn := startTestPostgres(t)
	seedSchema(t, dsn)

	// maxRows=2 is deliberately smaller than the 3 seeded rows, so we can
	// assert the truncation flag against a real query result.
	ds, err := postgres.New(context.Background(), dsn, 2)
	if err != nil {
		t.Fatalf("failed to connect DataSource: %v", err)
	}
	defer ds.Close()

	query := mustQuery(t, "SELECT id, name, email FROM customers ORDER BY id")

	result, err := ds.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error executing query: %v", err)
	}

	if result.RowCount != 2 {
		t.Fatalf("expected exactly 2 rows due to MaxRows, got %d", result.RowCount)
	}
	if !result.Truncated {
		t.Fatal("expected the result to be marked as truncated")
	}
}

func TestPostgresDataSource_ExecutePropagatesDatabaseErrors(t *testing.T) {
	dsn := startTestPostgres(t)

	ds, err := postgres.New(context.Background(), dsn, 0)
	if err != nil {
		t.Fatalf("failed to connect DataSource: %v", err)
	}
	defer ds.Close()

	// No schema was seeded, so this table genuinely doesn't exist — this
	// verifies that a real database error surfaces as a Go error rather
	// than being silently swallowed.
	query := mustQuery(t, "SELECT * FROM this_table_does_not_exist")

	if _, err := ds.Execute(context.Background(), query); err == nil {
		t.Fatal("expected an error when querying a nonexistent table, got none")
	}
}