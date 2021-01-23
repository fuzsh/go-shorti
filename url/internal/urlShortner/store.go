package urlShortner

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

	// CreateUser create new link.
	CreateLink(*sqlx.Tx, string, string) (int, error)

	// CreateUserLinkRelation create relation between user and link
	CreateUserLinkRelation(*sqlx.Tx, int, int) error
}
