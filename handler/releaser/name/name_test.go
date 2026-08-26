package name_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
)

func ExampleHumanize() {
	s, _ := name.Humanize("defacto2")
	fmt.Println(s)

	s, _ = name.Humanize("razor-1911-demo")
	fmt.Println(s)

	s, _ = name.Humanize("razor-1911-demo*trsi")
	fmt.Println(s)
	// Output:
	// defacto2
	// razor 1911 demo
	// razor 1911 demo, trsi
}

func ExampleHumanize_error() {
	_, err := name.Humanize("razor-1911-demo#trsi")
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// the path contains invalid characters
}

func ExampleNames() {
	find := name.Path("surprise-productions")
	for key, val := range name.Names() {
		if key == find {
			fmt.Println(val)
		}
	}
	// Output: Surprise! Productions
}

func ExampleObfuscate() {
	obf := name.Obfuscate("ACiD Productions")
	if !obf.Valid() {
		fmt.Println("invalid")
	} else {
		fmt.Println(string(obf))
	}
	// Output: acid-productions
}

func ExampleList() {
	uri := "defacto2net"
	for key, val := range name.Names() {
		if key == name.Path(uri) {
			fmt.Println(val)
		}
	}
	// Output: Defacto2 website
}

func ExampleUpper() {
	uri := "beer"
	for key, val := range name.Upper() {
		if key == name.Path(uri) {
			fmt.Println(val)
		}
	}
	// Output: BEER
}

func ExamplePath_String() {
	fmt.Println(name.Path("acid-productions").String())
	// Output: ACiD Productions
}

func ExamplePath_String_unlisted() {
	s := name.Path("defacto2").String()
	fmt.Println(len(s))
	// Output: 0
}

func ExamplePath_Valid() {
	fmt.Println(name.Path("defacto2").Valid())

	fmt.Println(name.Path("Defacto2").Valid())
	// Output: true
	// false
}

func TestNames(t *testing.T) {
	t.Parallel()
	// confirm all keys are valid and values are not empty
	for key, val := range name.Names() {
		// to debug, send to os.Stdout
		_, _ = fmt.Fprintln(io.Discard, key, val)
		if !key.Valid() {
			t.Errorf("Name() invalid %v", key)
		}
		if val == "" {
			t.Errorf("Name() empty value %v", key)
		}
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
			name:    "valid path",
			path:    "path/to/file",
			want:    "",
			wantErr: name.ErrInvalidPath,
		},
		{
			name:    "invalid path",
			path:    "",
			want:    "",
			wantErr: name.ErrInvalidPath,
		},
		{
			name:    "path with ampersand",
			path:    "path-ampersand-path",
			want:    "path & path",
			wantErr: nil,
		},
		{
			name:    "path with underscore",
			path:    "path_with_underscore",
			want:    "path-with-underscore",
			wantErr: nil,
		},
		{
			name:    "path with asterisk",
			path:    "path*with*asterisk",
			want:    "path, with, asterisk",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := name.Humanize(tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Humanize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Humanize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObfuscate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  string
		want string
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := name.Obfuscate(tt.arg)
			if got != name.Path(tt.want) {
				t.Errorf("Obfuscate(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}
