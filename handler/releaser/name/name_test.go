package name_test

import (
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  name.Path
	}{
		{"", ""},
		{"OldSkool", "oldskool"},
		{"Defacto2", ""},
		{"Intel", ""},
		{"rpgca", ""},
	}

	for n, tt := range tests {
		t.Run("testfind #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := name.Find(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestCopy(t *testing.T) {
	t.Parallel()

	// confirm all keys are valid and values are not empty
	for p, val := range name.Copy() {
		// to debug, send to os.Stdout
		_, _ = fmt.Fprintln(io.Discard, p, val)
		if !name.Valid(p) {
			t.Errorf("Name() invalid %v", p)
		}
		if val == "" {
			t.Errorf("Name() empty value %v", p)
		}
	}
}

func TestIndexes(t *testing.T) {
	t.Parallel()

	var key name.Path

	c := name.Copy()
	key = "pump"
	be.True(t, c[key] != "")

	l := name.Lower()
	key = "intel"
	be.True(t, l[key] != "")

	u := name.Upper()
	key = "rpgca"
	be.True(t, u[key] != "")

	a := name.All()
	key = "pump"
	be.True(t, a[key] != "")
	key = "intel"
	be.True(t, a[key] != "")
	key = "rpgca"
	be.True(t, a[key] != "")
}

func TestHumanize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path    name.Path
		want    string
		wantErr error
	}{
		{
			path:    "path/to/file",
			want:    "",
			wantErr: name.ErrPath,
		},
		{
			path:    "",
			want:    "",
			wantErr: name.ErrPath,
		},
		{
			path:    "path-ampersand-path",
			want:    "path & path",
			wantErr: nil,
		},
		{
			path:    "path_with_underscore",
			want:    "path-with-underscore",
			wantErr: nil,
		},
		{
			path:    "path*with*asterisk",
			want:    "path, with, asterisk",
			wantErr: nil,
		},
	}

	for n, tt := range tests {
		t.Run("test humanize #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got, err := name.Humanize(tt.path)
			be.Err(t, err, tt.wantErr)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestObfuscate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  string
		want name.Path
	}{
		{
			name: "empty string",
			arg:  "",
			want: "",
		},
		{
			name: "single word",
			arg:  "HeLlo",
			want: "hello",
		},
		{
			name: "multiple words",
			arg:  "Hello World",
			want: "hello-world",
		},
		{
			name: "ampersand",
			arg:  "Ben & Jerry's",
			want: "ben-ampersand-jerrys",
		},
		{
			name: "comma",
			arg:  "John, Paul, George, Ringo",
			want: "john*paul*george*ringo",
		},
		{
			name: "mixed",
			arg:  "The quick brown fox jumps over the lazy dog, but the dog is faster",
			want: "the-quick-brown-fox-jumps-over-the-lazy-dog*but-the-dog-is-faster",
		},
	}

	for n, tt := range tests {
		t.Run("test obfuscate #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := name.Obfuscate(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}
