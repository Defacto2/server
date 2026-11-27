package janeway_test

import (
	"testing"

	"github.com/Defacto2/server/handler/janeway"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		uri  string
		want janeway.GroupID
	}{
		{"empty", "", 0},
		{"unknown", "not-a-group", 0},
		{"case sensitive", "Fairlight", 0},
		{"fairlight", "fairlight", 252},
		{"razor 1911", "razor-1911", 338},
		{"trsi", "tristar-ampersand-red-sector-inc", 968},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := janeway.Find(tt.uri)
			be.Equal(t, got, tt.want)
		})
	}
}
