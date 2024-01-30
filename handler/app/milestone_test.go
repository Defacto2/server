package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestMilestone(t *testing.T) {
	t.Parallel()
	ms := app.Collection()
	assert.Equal(t, 108, ms.Len())
	assert.Equal(t, 108, len(ms))

	one := ms[0]
	assert.Equal(t, 1971, one.Year)
	assert.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		assert.NotEqual(t, 0, record.Year)
	}
}
