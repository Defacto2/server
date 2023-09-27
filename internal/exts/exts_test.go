package exts_test

import (
	"testing"

	"github.com/Defacto2/server/internal/exts"
	"github.com/stretchr/testify/assert"
)

func TestIcon(t *testing.T) {
	tests := []struct {
		name      string
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"", "", assert.Equal},
		{"myfile", "", assert.Equal},
		{"myimage.png", "image2", assert.Equal},
		{"double.exts.avi", "movie", assert.Equal},
		{"a web site .htm", "generic", assert.Equal},
		{"👾.mp3", "sound2", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, exts.IconName(tt.name))
		})
	}
}
