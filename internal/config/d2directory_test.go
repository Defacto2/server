package config_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestDirectory(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	file := filepath.Join(tmp, "file.txt")
	be.Err(t, os.WriteFile(file, []byte("x"), 0o600), nil)

	tests := []struct {
		name      string
		path      string
		wantErr   error
		wantIssue string
	}{
		{"empty", "", os.ErrNotExist, ""},
		{"missing", filepath.Join(tmp, "missing"), os.ErrNotExist, "not exist"},
		{"file", file, config.ErrNotDir, "points to a file"},
		{"valid", tmp, nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := config.Directory(tt.path)
			if tt.wantErr == nil {
				be.Err(t, d.Check(), nil)
			} else {
				be.True(t, errors.Is(d.Check(), tt.wantErr))
			}

			if tt.wantIssue == "" {
				be.Equal(t, d.Issue(), "")
			} else {
				be.True(t, strings.Contains(d.Issue(), tt.wantIssue))
			}

			be.Equal(t, d.String(), tt.path)
			be.Equal(t, d.LogValue().String(), tt.path)
		})
	}

	t.Run("permission", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}

		noAccess := filepath.Join(t.TempDir(), "locked")
		child := filepath.Join(noAccess, "sub")
		be.Err(t, os.MkdirAll(child, 0o700), nil)
		be.Err(t, os.Chmod(noAccess, 0o000), nil)
		t.Cleanup(func() { _ = os.Chmod(noAccess, 0o700) })

		d := config.Directory(child)
		be.True(t, errors.Is(d.Check(), fs.ErrPermission))
		be.True(t, strings.Contains(d.Issue(), "permission denied"))
	})
}
