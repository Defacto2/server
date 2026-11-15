package download

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nalgeon/be"
)

func TestTextContentType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"utf-16le", []byte{0xff, 0xfe, 'A', 0x00}, "text/plain; charset=utf-16le"},
		{"utf-16be", []byte{0xfe, 0xff, 0x00, 'A'}, "text/plain; charset=utf-16be"},
		{"latin1", []byte{0x8a, 'e', '\n'}, "text/plain; charset=iso-8859-1"},
		{"utf-8", []byte("hello world\n"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(dir, tt.name+".txt")
			err := os.WriteFile(p, tt.data, 0o600)
			be.Err(t, err, nil)

			got, err := textContentType(p)
			be.Err(t, err, nil)
			be.Equal(t, got, tt.want)
		})
	}
}
