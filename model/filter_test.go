package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

func TestSummCols(t *testing.T) {
	t.Parallel()

	cols := model.SummCols()
	be.Equal(t, cols, []string{
		postgres.SumSize,
		postgres.TotalCnt,
		postgres.MinYear,
		postgres.MaxYear,
	})
}

func TestFilterKey(t *testing.T) {
	t.Parallel()

	be.Equal(t, model.KeyNFO.String(), "nfo")
	be.True(t, model.KeyNFO.IsValid())
	be.True(t, model.KeyOneOffs.IsValid())
	be.True(t, model.KeyOneOffsAZ.IsValid())
	be.True(t, model.KeyWindowsPack.IsValid())
	be.True(t, !model.Key("").IsValid())
	be.True(t, !model.Key("not-a-filter").IsValid())
}
