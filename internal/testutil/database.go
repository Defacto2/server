//nolint:gochecknoglobals
package testutil

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/postgres"
)

// PostgreSQL database helpers

const pingTimeout = 3 * time.Second

var share = sync.OnceValues(func() (*sql.DB, error) {
	db, err := postgres.Open()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
})

// DB returns a shared Postgres database connection for use with tests.
// If no connection can be made, the test is skipped.
func DB(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := share()
	if db == nil {
		tb.Skipf("postgres database unavailable: %v", err)
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

	// Use context.Background() rather than tb.Context(): tb.Context() is canceled
	// before tb.Cleanup runs, which would cancel the transaction before tx.Rollback().
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		tb.Fatalf("begin tx: %v", err)
	}

	tb.Cleanup(func() {
		// returning errors for rollback is important due to the sensitivity
		// it has with contexts, timeouts, etc.
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			tb.Errorf("tx rollback: %v", err)
		}
	})

	return tx
}
