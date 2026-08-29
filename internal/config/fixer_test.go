package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestTestInfo(t *testing.T) {
	t.Parallel()

	buff := testutil.Buffer(t)

	config.TempInfo(buff.Log)

	be.True(t, buff.Contains("Temporary directory"))
	be.True(t, buff.Contains("Usage="))
}

func TestRecordCount(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	got := config.RecordCount(t.Context(), db)
	be.True(t, got > 2)
}
