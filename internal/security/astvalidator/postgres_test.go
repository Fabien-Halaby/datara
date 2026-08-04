package astvalidator_test

import (
	"testing"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
	"github.com/Fabien-Halaby/datara/internal/security/astvalidator"
)

func readOnlyValidator() *astvalidator.PostgresValidator {
	return astvalidator.NewPostgresValidator(domain.DefaultReadOnlyPolicy())
}

func TestValidate_AcceptsSimpleSelect(t *testing.T) {
	v := readOnlyValidator()

	query, err := v.Validate("SELECT id, name FROM customers LIMIT 10")
	if err != nil {
		t.Fatalf("expected a plain SELECT to be accepted, got error: %v", err)
	}
	if query.Raw() == "" {
		t.Fatal("expected a non-empty validated query")
	}
}

func TestValidate_AcceptsSelectWithJoinAndAggregation(t *testing.T) {
	v := readOnlyValidator()

	sql := `SELECT c.name, COUNT(o.id) FROM customers c
	        JOIN orders o ON o.customer_id = c.id
	        GROUP BY c.name`
	if _, err := v.Validate(sql); err != nil {
		t.Fatalf("expected a SELECT with JOIN/GROUP BY to be accepted, got error: %v", err)
	}
}

func TestValidate_RejectsWriteStatements(t *testing.T) {
	v := readOnlyValidator()

	writes := []string{
		"DELETE FROM orders",
		"DELETE FROM orders WHERE status = 'cancelled'",
		"UPDATE customers SET name = 'x'",
		"INSERT INTO customers (name) VALUES ('x')",
		"DROP TABLE orders",
		"TRUNCATE orders",
		"ALTER TABLE orders ADD COLUMN foo TEXT",
		"CREATE TABLE evil (id INT)",
	}

	for _, sql := range writes {
		t.Run(sql, func(t *testing.T) {
			if _, err := v.Validate(sql); err == nil {
				t.Fatalf("expected %q to be rejected, but it was accepted", sql)
			}
		})
	}
}

func TestValidate_RejectsStackedStatements(t *testing.T) {
	v := readOnlyValidator()

	// Classic stacked-query injection attempt: a legit-looking SELECT
	// hiding a destructive statement behind a semicolon.
	sql := "SELECT * FROM customers; DROP TABLE customers;"

	if _, err := v.Validate(sql); err == nil {
		t.Fatal("expected stacked statements to be rejected, but the query was accepted")
	}
}

func TestValidate_RejectsSelectInto(t *testing.T) {
	v := readOnlyValidator()

	// SELECT ... INTO silently creates a new table — a side effect that
	// must never be allowed under a read-only policy.
	sql := "SELECT * INTO new_table FROM customers"

	if _, err := v.Validate(sql); err == nil {
		t.Fatal("expected SELECT ... INTO to be rejected, but the query was accepted")
	}
}

func TestValidate_RejectsLockingClauses(t *testing.T) {
	v := readOnlyValidator()

	locking := []string{
		"SELECT * FROM orders FOR UPDATE",
		"SELECT * FROM orders FOR SHARE",
	}

	for _, sql := range locking {
		t.Run(sql, func(t *testing.T) {
			if _, err := v.Validate(sql); err == nil {
				t.Fatalf("expected %q to be rejected, but it was accepted", sql)
			}
		})
	}
}

func TestValidate_RejectsEmptyQuery(t *testing.T) {
	v := readOnlyValidator()

	if _, err := v.Validate("   "); err == nil {
		t.Fatal("expected an empty query to be rejected")
	}
}

func TestValidate_RejectsInvalidSQL(t *testing.T) {
	v := readOnlyValidator()

	if _, err := v.Validate("SELEKT * FROM customers"); err == nil {
		t.Fatal("expected malformed SQL to be rejected")
	}
}

func TestValidate_EnforcesTableAllowlist(t *testing.T) {
	policy := domain.SecurityPolicy{
		ReadOnly:      true,
		AllowedTables: []string{"customers"},
		MaxRows:       100,
	}
	v := astvalidator.NewPostgresValidator(policy)

	if _, err := v.Validate("SELECT * FROM customers"); err != nil {
		t.Fatalf("expected access to an allowed table to succeed, got error: %v", err)
	}

	if _, err := v.Validate("SELECT * FROM orders"); err == nil {
		t.Fatal("expected access to a non-allowlisted table to be rejected")
	}
}

func TestValidate_EmptyAllowlistMeansUnrestricted(t *testing.T) {
	// A policy with no AllowedTables set should not restrict access to
	// any table at all — this documents the "empty means unrestricted"
	// behavior so it doesn't get silently broken by a future change.
	policy := domain.SecurityPolicy{ReadOnly: true, MaxRows: 100}
	v := astvalidator.NewPostgresValidator(policy)

	if _, err := v.Validate("SELECT * FROM anything_at_all"); err != nil {
		t.Fatalf("expected unrestricted policy to allow any table, got error: %v", err)
	}
}