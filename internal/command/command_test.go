package command_test

import (
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/stretchr/testify/assert"
)

func TestBaseName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"No extension", "/path/to/file", "file"},
		{"With extension", "/path/to/file.txt", "file"},
		{"Multiple extensions", "/path/to/file.tar.gz", "file.tar"},
		{"Hidden file", "/path/to/.hidden", ""},
		{"Hidden file with extension", "/path/to/.hidden.txt", ".hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, command.BaseName(tt.path))
		})
	}
}

func TestBaseNamePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"No extension", "/path/to/file", "/path/to/file"},
		{"With extension", "/path/to/file.txt", "/path/to/file"},
		{"Multiple extensions", "/path/to/file.tar.gz", "/path/to/file.tar"},
		{"Hidden file", "/path/to/.hidden", "/path/to"},
		{"Hidden file with extension", "/path/to/.hidden.txt", "/path/to/.hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, command.BaseNamePath(tt.path))
		})
	}
}
