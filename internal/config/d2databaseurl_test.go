package config_test

import (
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestDatabaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"no user", "postgres://localhost:5432/db", "postgres://localhost:5432/db"},
		{"user only", "postgres://root@localhost:5432/db", "postgres://root@localhost:5432/db"},
		{"password", "postgres://root:secret@localhost:5432/db?sslmode=disable", "postgres://root:xxxxx@localhost:5432/db?sslmode=disable"},
		{"invalid", "postgres://root:secret@localhost:badport/db", "xxxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := config.Connection(tt.raw)
			be.Equal(t, c.String(), tt.want)
			be.Equal(t, c.LogValue().String(), tt.want)
		})
	}
}
