package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestRecordsSub(t *testing.T) {
	t.Parallel()
	s := app.RecordsSub("")
	assert.Equal(t, "unknown uri", s)

	// range to 57
	for i := 1; i <= int(57); i++ {
		assert.NotEqual(t, "unknown uri", app.URI(i).String())
	}
}
