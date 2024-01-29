package helper_test

import (
	"testing"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestLines(t *testing.T) {
	i, err := helper.Lines("")
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("nosuchfile")
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("../testdata")
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("../testdata/TEST.BMP")
	assert.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("../testdata/TESTS.TXT")
	assert.NoError(t, err)
	assert.Equal(t, 4, i)
}
