package lism_test

import (
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/nalgeon/be"
)

func TestAll(t *testing.T) {
	t.Parallel()

	l := lism.All()
	be.True(t, len(l) >= 100)

	found := false
	for _, vals := range l {
		if slices.ContainsFunc(vals, func(v string) bool {
			return strings.EqualFold(v, "inc")
		}) {
			found = true
			break
		}
	}
	be.True(t, found)
}

func TestExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path lism.Path
		want bool
	}{
		{"empty", "", false},
		{"unknown", "some-random-bbs", false},
		{"known", "tristar", true},
		{"multiple", "tristar-ampersand-red-sector-inc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.Exist(tt.path)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lism.Path
	}{
		{"empty", "", nil},
		{"no match", "some-unknown-random-bbs", nil},
		{"df2", "df2", []lism.Path{"defacto2", "defacto2net"}},
		{"razor", "RzR", []lism.Path{"razor-1911", "razor-1911-demo", "razordox"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Clone(lism.Find(tt.input))
			slices.Sort(got)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFindOne(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  lism.Path
	}{
		{"empty", "", ""},
		{"no match", "some-unknown-random-bbs", ""},
		{"unique", "FiRM", "the-firm"},
		{"case insensitive", "nappa", "north-american-pirate_phreak-association"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.FindOne(tt.input)
			be.Equal(t, got, tt.want)
		})
	}

	t.Run("collision", func(t *testing.T) {
		t.Parallel()

		got := lism.FindOne("df2")
		be.True(t, slices.Contains([]lism.Path{"defacto2", "defacto2net"}, got))
	})
}

func TestInitialism(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path lism.Path
		want []string
	}{
		{"empty", "", nil},
		{"unknown", "some-random-bbs", nil},
		{"known", "union", []string{"UNi"}},
		{"multiple", "wave", []string{"The Wave", "CNC"}},
		{"df2", "defacto2", []string{"DF2", "DF"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.Initialism(tt.path)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	for key, vals := range lism.All() {
		k := string(key)
		if k == "" {
			t.Error("empty key in initialisms")
			continue
		}
		chr := rune(k[0])
		if k != strings.ToLower(k) {
			t.Errorf("%q: key must be lowercase", k)
		}
		if k != strings.TrimSpace(k) {
			t.Errorf("%q: key has leading or trailing whitespace", k)
		}
		if !unicode.IsLetter(chr) && !unicode.IsNumber(chr) {
			t.Errorf("%q: key must start with a letter or number", k)
		}
		if len(vals) == 0 {
			t.Errorf("%q: empty initialisms slice", k)
		}
		for _, v := range vals {
			if v == "" {
				t.Errorf("%q: empty initialism value", k)
			}
			if v != strings.TrimSpace(v) {
				t.Errorf("%q: initialism %q has leading or trailing whitespace", k, v)
			}
		}
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path lism.Path
		want string
	}{
		{"empty", "", ""},
		{"unknown", "some-random-bbs", ""},
		{"single", "union", "UNi"},
		{"multiple", "defacto2", "DF2, DF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.String(tt.path)
			be.Equal(t, got, tt.want)
		})
	}
}
