package csdb_test

import (
	"testing"

	"github.com/Defacto2/server/handler/csdb"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		uri  string
		want csdb.GroupID
	}{
		{"empty", "", 0},
		{"unknown", "not-a-group", 0},
		{"case sensitive", "Fairlight", 0},
		{"fairlight", "fairlight", 20},
		{"razor 1911", "razor-1911", 431},
		{"trsi", "tristar-ampersand-red-sector-inc", 915},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := csdb.Find(tt.uri)
			be.Equal(t, got, tt.want)
		})
	}
}
