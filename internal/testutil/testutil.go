package testutil

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/Defacto2/server/internal/postgres"
)

var share = sync.OnceValue(func() *sql.DB {
	db, err := postgres.Open()
	if err != nil {
		return nil
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil
	}
	return db
})

// DB returns a shared Postgres database connection for use with tests.
// If no connection can be made, the test is skipped.
func DB(tb testing.TB) *sql.DB {
	tb.Helper()

	db := share()
	if db == nil {
		tb.SkipNow()
	}
	return db
}

// Tx returns a shared Postgres database connection for use with tests.
// The transaction with any database edits get rolled back during the test cleanup.
// If no connection can be made, the test is skipped.
//
// Note: Only one Tx should be used per-test otherwise the test may never resolve.
func Tx(tb testing.TB) *sql.Tx {
	tb.Helper()

	db := DB(tb)

	tx, err := db.Begin()
	if err != nil {
		tb.Skip("tx failed")
	}

	tb.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			tb.Errorf("tx rollback: %v", err)
		}
	})

	return tx
}

// Connections logs the active and maximum Postgres database connections.
func Connections(tb testing.TB) {
	tb.Helper()

	db := share()
	if db == nil {
		return
	}
	a, m, err := postgres.Connections(db)
	if err != nil {
		tb.Fatalf("connections err: %v", err)
	}
	tb.Log("active database connections:", a, "maximum:", m)
}
