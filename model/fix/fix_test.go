package fix_test

import (
	"testing"

	"github.com/Defacto2/server/model/fix"
	"github.com/stretchr/testify/assert"
)

func TestCFToUUIDv1(t *testing.T) {
	cfid := "00000000-0000-0000-0000000000000000"
	expected := "00000000-0000-0000-0000-000000000000"
	result, err := fix.CFToUUIDv1(cfid)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
