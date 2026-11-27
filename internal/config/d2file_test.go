package config_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testdata = testutil.FileData("implodezip")

func TestFile(t *testing.T) {
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
		{"dir", tmp, config.ErrNotFile, "points to a directory"},
		{"valid", file, nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := config.File(tt.path)
			if tt.wantErr == nil {
				be.Err(t, f.Check(), nil)
			} else {
				be.True(t, errors.Is(f.Check(), tt.wantErr))
			}

			if tt.wantIssue == "" {
				be.Equal(t, f.Issue(), "")
			} else {
				be.True(t, strings.Contains(f.Issue(), tt.wantIssue))
			}

			be.Equal(t, f.String(), tt.path)
			be.Equal(t, f.LogValue().String(), tt.path)
		})
	}

	t.Run("permission", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}

		noAccess := filepath.Join(t.TempDir(), "locked")
		child := filepath.Join(noAccess, "secret.txt")
		be.Err(t, os.Mkdir(noAccess, 0o700), nil)
		be.Err(t, os.WriteFile(child, []byte("x"), 0o600), nil)
		be.Err(t, os.Chmod(noAccess, 0o000), nil)
		t.Cleanup(func() { _ = os.Chmod(noAccess, 0o700) })

		f := config.File(child)
		be.True(t, errors.Is(f.Check(), fs.ErrPermission))
		be.True(t, strings.Contains(f.Issue(), "permission denied"))
	})
}
