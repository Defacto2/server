package helper_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/server/internal/exts"
	"github.com/Defacto2/server/internal/helper"
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
			tt.assertion(t, tt.expect, helper.Finds(tt.args.name, tt.args.names...))
		})
	}
}

func TestIsFile(t *testing.T) {
	self := filepath.Join(".", "bool_test.go")
	tests := []struct {
		name      string
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{self, true, assert.Equal},
		{"^&%#$%@#", false, assert.Equal},
		{"testdata/", false, assert.Equal},
		{"testdata/TEST.DOC", true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.IsFile(tt.name))
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
		{"testdata/", true, assert.Equal},
		{"testdata/TEST.DOC", true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.IsStat(tt.name))
		})
	}
}

func TestBools(t *testing.T) {
	assert.False(t, helper.IsDay(-1))
	assert.False(t, helper.IsDay(32))
	assert.True(t, helper.IsDay(1))
	assert.False(t, helper.IsYear(-1))
	assert.True(t, helper.IsYear(1970))
	assert.False(t, helper.IsYear(time.Now().Year()+1))
	assert.False(t, exts.IsApp("myapp"))
	assert.True(t, exts.IsApp("myapp.exe"))
	assert.True(t, exts.IsArchive("stuff.zip"))
	assert.True(t, exts.IsDocument("readme.doc"))
	assert.True(t, exts.IsImage("cat.jpeg"))
	assert.True(t, exts.IsHTML("index.html"))
	assert.True(t, exts.IsAudio("song.wav"))
	assert.True(t, exts.IsTune("song.mod"))
	assert.True(t, exts.IsVideo("cat.divx"))
}
