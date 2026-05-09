package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Querier abstracts database operations shared by *sqlx.DB and *sqlx.Tx.
// This implements the Strategy Pattern + Liskov Substitution Principle:
// any function accepting Querier works identically with a raw connection or a transaction.
//
// Usage:
//   func (s *MyService) DoWork(ctx context.Context, q Querier) error {
//       return q.GetContext(ctx, &result, "SELECT ...", args...)
//   }
//
// Callers pass either db or tx:
//   s.DoWork(ctx, db)   // direct
//   s.DoWork(ctx, tx)   // inside transaction
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Compile-time verification that both types satisfy Querier.
var _ Querier = (*sqlx.DB)(nil)
var _ Querier = (*sqlx.Tx)(nil)
