package initialism_test

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/Defacto2/server/handler/releaser/initialism"
	"github.com/nalgeon/be"
)

func ExampleInitialism() {
	fmt.Println(initialism.Initialism("defacto2"))
	// Output: [DF2 DF]
}

func ExampleInitialisms() {
	const find = "USA"
	for key, isms := range *initialism.Initialisms() {
		if slices.Contains(isms, find) {
			fmt.Printf("Found %v in %v\n", find, key)
		}
	}
	// Output: Found USA in united-software-association*fairlight
}

func ExampleIsInitialism() {
	fmt.Println(initialism.IsInitialism("defacto2"))
	// Output: true
}

func ExampleJoin() {
	fmt.Println(initialism.Join("the-firm")) // FiRM, FRM

	fmt.Println(initialism.Join("united-software-association*fairlight")) // USA
	// Output: FiRM, FRM
	// USA/Fairlight, USA/FLT, USA
}

func TestMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"empty", "", []string{}},
		{"no match", "some-unknown-random-bbs", []string{}},
		{"df2", "df2", []string{"defacto2", "defacto2net"}},
		{"razor", "RzR", []string{"razor-1911", "razor-1911-demo", "razordox"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := initialism.Match(tt.s)
			c := make([]string, len(got))
			for i, v := range got {
				c[i] = string(v)
			}
			sort.Strings(c)
			be.Equal(t, c, tt.want)
			// if !assert.Equal(t, tt.want, c) {
			// 	t.Errorf("Match() = %v, want %v", c, tt.want)
			// }
		})
	}
}

func BenchmarkIsInitialism(b *testing.B) {
	for b.Loop() {
		fmt.Fprintln(io.Discard, initialism.IsInitialism("defacto2"))
	}
}

func BenchmarkInitialism(b *testing.B) {
	for b.Loop() {
		fmt.Fprintln(io.Discard, initialism.Initialism("defacto2"))
	}
}

func BenchmarkInitialisms(b *testing.B) {
	for b.Loop() {
		const find = "USA"
		for key, values := range *initialism.Initialisms() {
			for value := range slices.Values(values) {
				if value == find {
					fmt.Fprintf(io.Discard, "Found %v in %v\n", find, key)
					return
				}
			}
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	for b.Loop() {
		fmt.Fprint(io.Discard, initialism.Match("razor"))
	}
}

func TestInitialism(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path initialism.Path
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
			if got := initialism.Initialism(tt.path); !equal(got, tt.want) {
				t.Errorf("Initialism() = %v, want %v", got, tt.want)
			}
		})
	}
	// Confirm all keys are valid URL paths.
	for key := range *initialism.Initialisms() {
		// keys must be lowercase and start with only letters or numbers
		k := string(key)
		chr := rune(k[0])
		be.Equal(t, k, strings.ToLower(k))
		be.Equal(t, k, strings.TrimSpace(k))
		valid := unicode.IsLetter(chr) || unicode.IsNumber(chr)
		be.True(t, valid)
	}
}

func TestInitialisms(t *testing.T) {
	t.Parallel()
	l := *initialism.Initialisms()
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
		path initialism.Path
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
			if got := initialism.IsInitialism(tt.path); got != tt.want {
				t.Errorf("IsInitialism() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
