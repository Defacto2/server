package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestInterviewees(t *testing.T) {
	t.Parallel()
	i := app.Interviewees()
	l := len(i)
	const expected = 12
	be.Equal(t, l, expected)

	for _, x := range i {
		be.True(t, x.Name != "")
	}
}
