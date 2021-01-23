package track

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// NewTenantID is a helper function to return a sql.NullInt64.
// The ID is considered valid if greater than 0.
func NewTenantID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}

// NullTenant can be used to pass no (null) tenant to filters and functions.
// This is a sql.NullInt64 with a value of 0.
var NullTenant = NewTenantID(0)

// Store defines an interface to persists hits and other data.
type Store interface {
	// NewTx creates a new transaction and panic on failure.
	NewTx() *sqlx.Tx

	// Commit commits given transaction and logs the error.
	Commit(*sqlx.Tx)

	// Rollback rolls back given transaction and logs the error.
	Rollback(*sqlx.Tx)

	// SaveHits persists a list of hits.
	SaveHits([]Hit) error
}