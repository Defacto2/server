package lism_test

import (
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/Defacto2/server/handler/releaser/lism"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"no match", "some-unknown-random-bbs", []string{}},
		{"df2", "df2", []string{"defacto2", "defacto2net"}},
		{"razor", "RzR", []string{"razor-1911", "razor-1911-demo", "razordox"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.Find(tt.input)
			c := make([]string, len(got))
			for i, v := range got {
				c[i] = string(v)
			}
			sort.Strings(c)
			be.Equal(t, c, tt.want)
		})
	}
}

func TestInitialism(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path lism.Path
		want []string
	}{
		{"empty path", "", nil},
		{"unknown path", "some-random-bbs", nil},
		{"known", "union", []string{"UNi"}},
		{"multiple", "wave", []string{"The Wave", "CNC"}},
		{"df2", "defacto2", []string{"DF2", "DF"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lism.Initialism(tt.path)
			be.True(t, matchSlice(t, got, tt.want))
		})
	}

	// Confirm all keys are valid URL paths.
	for key := range lism.Copy() {
		// keys must be lowercase and start with only letters or numbers
		k := string(key)
		chr := rune(k[0])
		be.Equal(t, k, strings.ToLower(k))
		be.Equal(t, k, strings.TrimSpace(k))
		valid := unicode.IsLetter(chr) || unicode.IsNumber(chr)
		be.True(t, valid)
	}
}

func TestCopy(t *testing.T) {
	t.Parallel()

	l := lism.Copy()
	if len(l) == 0 {
		t.Errorf("Initialisms() = %v, want %v", l, "non-empty")
	}
	if len(l) < 100 {
		t.Errorf("Initialisms() = %v, want %v", l, "more than 100")
	}

	s := "inc"
	m := ""
	for _, v := range l {
		for _, x := range v {
			if strings.ToLower(x) == s {
				m = x
			}
		}
	}
	if m == "" {
		t.Errorf("Initialisms() could not find %v", s)
	}
}

func TestIsInitialism(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path lism.Path
		want bool
	}{
		{"empty path", "", false},
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

func matchSlice(t *testing.T, got, want []string) bool {
	t.Helper()

	if len(got) != len(want) {
		t.Log(len(got), len(want))
		return false
	}

	for i, v := range got {
		if v != want[i] {
			t.Log(v, want)
			return false
		}
	}

	return true
}
