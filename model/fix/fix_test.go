package fix_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

func TestMagics(t *testing.T) {
	// when testing, go may cache the test result after the first run
	t.Parallel()
	db, err := postgres.Open()
	be.Err(t, err, nil)
	defer func() {
		if err := db.Close(); err != nil {
			be.Err(t, err, nil)
		}
	}()
	if err := db.Ping(); err != nil {
		// skip the test if the database is not available
		return
	}
	err = fix.Magics(db)
	be.Err(t, err, nil)
}
