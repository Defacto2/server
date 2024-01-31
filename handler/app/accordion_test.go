package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()
	list := app.List()
	assert.Len(t, list, 9)
}
