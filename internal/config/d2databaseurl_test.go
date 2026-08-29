package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/nalgeon/be"
)

func TestDatabaseURL(t *testing.T) {
	t.Parallel()

	const testurl = postgres.DefaultURL
	const mask = "postgres://root:xxxxx@"

	got := config.Connection(testurl).String()
	be.True(t, strings.HasPrefix(got, mask))
}
