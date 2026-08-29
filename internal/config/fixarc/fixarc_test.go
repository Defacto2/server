package fixarc_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestFiles(t *testing.T) {
	t.Parallel()

	_, err := fixarc.Files(t.Context(), nil)
	be.Err(t, err)

	db := testutil.DB(t)
	fs, err := fixarc.Files(t.Context(), db)
	be.Err(t, err, nil)
	be.True(t, len(fs) > 0)
}
