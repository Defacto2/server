package readme_test

import (
	"testing"

	"github.com/Defacto2/server/handler/readme"
	"github.com/nalgeon/be"
)

func TestAddPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		prefix string
		want   string
	}{
		{"both", "hello world", "prefix", "prefix\n\nhello world"},
		{"empty prefix", "hello", "", "hello"},
		{"whitespace prefix", "hello", " \t\n ", "hello"},
		{"empty body", "", "prefix", "prefix"},
		{"whitespace body", " \t\n ", "prefix", "prefix"},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := readme.AddPrefix([]byte(tt.body), []byte(tt.prefix))
			be.Equal(t, string(got), tt.want)
		})
	}
}

func TestAddSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		suffix string
		want   string
	}{
		{"both", "hello world", "suffix", "hello world\n\nsuffix"},
		{"empty suffix", "hello", "", "hello"},
		{"whitespace suffix", "hello", " \t\n ", "hello"},
		{"empty body", "", "suffix", "suffix"},
		{"whitespace body", " \t\n ", "suffix", "suffix"},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := readme.AddSuffix([]byte(tt.body), []byte(tt.suffix))
			be.Equal(t, string(got), tt.want)
		})
	}
}
