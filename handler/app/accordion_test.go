package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestList(t *testing.T) {
	t.Parallel()

	list := app.List()
	const expectedCount = 9
	be.True(t, len(list) == expectedCount)
	be.Equal(t, list[0].Name, "Text art scene")
}
