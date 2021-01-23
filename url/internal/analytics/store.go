package analytics

import (
	"github.com/jmoiron/sqlx"
)

// Store defines an interface to persists hits and other data.
type Store interface {
	// NewTx creates a new transaction and panic on failure.
	NewTx() *sqlx.Tx

	// Commit commits given transaction and logs the error.
	Commit(*sqlx.Tx)

	// Rollback rolls back given transaction and logs the error.
	Rollback(*sqlx.Tx)

	// CreateUser create new user.
	GetAnalytics(*sqlx.Tx, Config, int) (interface{}, error)
}
