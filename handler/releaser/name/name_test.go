package name_test

import (
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
	"github.com/nalgeon/be"
)

func TestAll(t *testing.T) {
	t.Parallel()

	// confirm all keys across names, lower, and upper are valid and values are not empty
	for p, val := range name.All() {
		if !name.Valid(p) {
			t.Errorf("All() invalid %v", p)
		}
		if val == "" {
			t.Errorf("All() empty value %v", p)
		}
	}
}

func TestClean(t *testing.T) {
	t.Parallel()

	isAlpha := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty", "", ""},
		{"unchanged", "Defacto", "Defacto"},
		{"strip digits and symbols", "Defacto2! 1996", "Defacto"},
		{"strip all", "12345!@#", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := name.Clean(tt.arg, isAlpha)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  name.Path
	}{
		{"empty", "", ""},
		{"styled", "OldSkool", "oldskool"},
		{"unlisted", "Defacto2", ""},
		{"lowercase", "Intel", ""},
		{"uppercase", "rpgca", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := name.Find(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestHumanize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    name.Path
		want    string
		wantErr error
	}{
		{
			name:    "invalid slash",
			path:    "path/to/file",
			want:    "",
			wantErr: name.ErrPath,
		},
		{
			name:    "empty",
			path:    "",
			want:    "",
			wantErr: name.ErrPath,
		},
		{
			name:    "ampersand",
			path:    "path-ampersand-path",
			want:    "path & path",
			wantErr: nil,
		},
		{
			name:    "underscore",
			path:    "path_with_underscore",
			want:    "path-with-underscore",
			wantErr: nil,
		},
		{
			name:    "asterisk",
			path:    "path*with*asterisk",
			want:    "path, with, asterisk",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := name.Humanize(tt.path)
			be.Err(t, err, tt.wantErr)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestIndexes(t *testing.T) {
	t.Parallel()

	c := name.Copy()
	be.Equal(t, c["pump"], "Pump!")

	l := name.Lower()
	be.Equal(t, l["intel"], "intel")

	u := name.Upper()
	be.Equal(t, u["rpgca"], "RPGCA")

	a := name.All()
	be.Equal(t, a["pump"], "Pump!")
	be.Equal(t, a["intel"], "intel")
	be.Equal(t, a["rpgca"], "RPGCA")
	be.Equal(t, len(a), len(c)+len(l)+len(u))
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
			name: "hyphen",
			arg:  "TDU-Jam!",
			want: "tdu_jam",
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := name.Obfuscate(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path name.Path
		want string
	}{
		{"empty", "", ""},
		{"unlisted", "defacto2", ""},
		{"styled", "acid-productions", "ACiD Productions"},
		{"case insensitive", "ACID-PRODUCTIONS", "ACiD Productions"},
		{"lowercase", "intel", "intel"},
		{"uppercase", "rpgca", "RPGCA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := name.String(tt.path)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path name.Path
		want bool
	}{
		{"empty", "", false},
		{"lowercase and digits", "razor-1911", true},
		{"allowed punctuation", "a-b_c*d&0", true},
		{"uppercase", "Defacto2", false},
		{"space", "razor 1911", false},
		{"slash", "path/to/file", false},
		{"hash", "razor-1911#trsi", false},
		{"bang", "acid-productions!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := name.Valid(tt.path)
			be.Equal(t, got, tt.want)
		})
	}
}
