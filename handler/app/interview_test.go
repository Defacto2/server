package app_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/stretchr/testify/assert"
)

func TestInterviewees(t *testing.T) {
	t.Parallel()
	i := app.Interviewees()
	assert.Equal(t, 9, len(i))

	for _, x := range i {
		assert.NotEmpty(t, x.Name)
	}
}
