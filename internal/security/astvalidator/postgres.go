package astvalidator

import (
	"fmt"

	pgquery "github.com/pganalyze/pg_query_go/v5"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

// PostgresValidator enforces a domain.SecurityPolicy against Postgres SQL
// by parsing each query into its AST via libpg_query (through
// pg_query_go) the same parser Postgres itself uses. This is what makes
// the read-only guarantee robust: Datara isn't pattern-matching keywords
// in a string, it's reading the actual parse tree Postgres would produce.
type PostgresValidator struct{
	policy domain.SecurityPolicy
}

// NewPostgresValidator builds a validator that enforces the given policy.
func NewPostgresValidator(policy domain.SecurityPolicy) *PostgresValidator {
	return &PostgresValidator{policy: policy}
}

// Validate parses raw and rejects anything that is not a single,
// side-effect-free SELECT statement.
func (v *PostgresValidator) Validate(raw string) (domain.SQLQuery, error) {
	query := domain.NewSQLQuery(raw)
	if query.IsZero() {
		return domain.SQLQuery{}, domain.ErrEmptyQuery
	}

	result, err := pgquery.Parse(query.Raw())
	if err != nil {
		return domain.SQLQuery{}, err
	}

	if len(result.Stmts) != 1 {
		return domain.SQLQuery{}, fmt.Errorf("only a single statement is allowed per query")
	}

	selectStmt := result.Stmts[0].GetSelectStmt()
	if selectStmt == nil {
		return domain.SQLQuery{}, fmt.Errorf("only SELECT statements are allowed")
	}

	if v.policy.ReadOnly {
		if selectStmt.IntoClause != nil {
			return domain.SQLQuery{}, fmt.Errorf("SELECT ... INTO is not allowed (creates a table)")
		}
		if len(selectStmt.LockingClause) > 0 {
			return domain.SQLQuery{}, fmt.Errorf("locking clauses (FOR UPDATE/SHARE) are not allowed")
		}
	}

	if len(v.policy.AllowedTables) > 0 {
		for _, table := range extractTableNames(selectStmt) {
			if !v.policy.IsTableAllowed(table) {
				return domain.SQLQuery{}, fmt.Errorf("access to table %q is not allowed", table)
			}
		}
	}

	return query, nil
}

// extractTableNames walks the FROM clause of a SELECT statement and
// returns the table names it references. This is a best-effort v1
// implementation: it covers plain FROM/JOIN targets, not every possible
// range expression (subqueries and CTEs are not resolved to table names
// here a future iteration can walk WithClause too).
func extractTableNames(stmt *pgquery.SelectStmt) []string {
	var tables []string
	for _, node := range stmt.FromClause {
		if rv := node.GetRangeVar(); rv != nil {
			tables = append(tables, rv.Relname)
		}
	}

	return tables
}