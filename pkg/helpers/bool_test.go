package helpers_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/stretchr/testify/assert"
)

func TestFinds(t *testing.T) {
	s := []string{"abc", "def", "ghi"}
	type args struct {
		name  string
		names []string
	}
	tests := []struct {
		args      args
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{args{"", nil}, false, assert.Equal},
		{args{"", []string{}}, false, assert.Equal},
		{args{"xyz", s}, false, assert.Equal},
		{args{"def", s}, true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.args.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helpers.Finds(tt.args.name, tt.args.names...))
		})
	}
}

func TestIsStat(t *testing.T) {
	self := filepath.Join(".", "bool_test.go")
	tests := []struct {
		name      string
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{self, true, assert.Equal},
		{"^&%#$%@#", false, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helpers.IsStat(tt.name))
		})
	}
}

func TestBools(t *testing.T) {
	assert.False(t, helpers.IsDay(-1))
	assert.False(t, helpers.IsDay(32))
	assert.True(t, helpers.IsDay(1))
	assert.False(t, helpers.IsYear(-1))
	assert.True(t, helpers.IsYear(1970))
	assert.False(t, helpers.IsYear(time.Now().Year()+1))
	assert.False(t, helpers.IsApp("myapp"))
	assert.True(t, helpers.IsApp("myapp.exe"))
	assert.True(t, helpers.IsArchive("stuff.zip"))
	assert.True(t, helpers.IsDocument("readme.doc"))
	assert.True(t, helpers.IsImage("cat.jpeg"))
	assert.True(t, helpers.IsHTML("index.html"))
	assert.True(t, helpers.IsAudio("song.wav"))
	assert.True(t, helpers.IsTune("song.mod"))
	assert.True(t, helpers.IsVideo("cat.divx"))
}
