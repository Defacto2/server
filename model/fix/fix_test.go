package fix_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model/fix"
	"github.com/stretchr/testify/require"
)

func TestMagics(t *testing.T) {
	// when testing, go may cache the test result after the first run
	t.Parallel()
	db, err := postgres.Open()
	require.NoError(t, err)
	defer db.Close()

	if err := db.Ping(); err != nil {
		// skip the test if the database is not available
		return
	}

	err = fix.Magics(db)
	require.NoError(t, err)
}
