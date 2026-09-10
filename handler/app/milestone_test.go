package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestMilestone(t *testing.T) {
	t.Parallel()

	ms := app.Collection()

	const expectedMileStones = 100
	be.True(t, ms.Count() > expectedMileStones)

	one := ms[0]
	const expectedYear = 1971
	be.Equal(t, expectedYear, one.Year)
	be.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		be.True(t, record.Year != 0)
	}
}
