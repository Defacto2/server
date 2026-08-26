package model_test

import (
	"testing"

	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
)

// Test that getColumns() returns cached columns on subsequent calls.
func TestGetColumnsCaching(t *testing.T) {
	t.Parallel()

	// populate cache 1
	call1 := model.GetColumns()
	be.True(t, call1 != nil)
	be.Equal(t, len(call1), 4)

	// populate cache 2
	call2 := model.GetColumns()
	be.True(t, call2 != nil)
	be.Equal(t, len(call1), len(call2))
}
