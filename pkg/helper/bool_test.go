package helper_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Defacto2/server/pkg/helper"
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
	assert.False(t, helper.IsApp("myapp"))
	assert.True(t, helper.IsApp("myapp.exe"))
	assert.True(t, helper.IsArchive("stuff.zip"))
	assert.True(t, helper.IsDocument("readme.doc"))
	assert.True(t, helper.IsImage("cat.jpeg"))
	assert.True(t, helper.IsHTML("index.html"))
	assert.True(t, helper.IsAudio("song.wav"))
	assert.True(t, helper.IsTune("song.mod"))
	assert.True(t, helper.IsVideo("cat.divx"))
}
