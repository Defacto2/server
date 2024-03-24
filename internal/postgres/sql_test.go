package postgres_test

import (
	"testing"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_Query(t *testing.T) {
	var v postgres.Version
	err := v.Query()
	require.Error(t, err)
	assert.Empty(t, v)
}

func TestVersion_String(t *testing.T) {
	var v postgres.Version
	assert.Empty(t, v.String())
}

func TestRole_Role(t *testing.T) {
	r := postgres.Roles()
	assert.NotEmpty(t, r)
}

func TestRole_Select(t *testing.T) {
	var r postgres.Role
	s := r.Distinct()
	assert.Contains(t, s, "SELECT DISTINCT")
}
