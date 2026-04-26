package fulltext_test

import (
	"embed"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/nalgeon/be"
)

func TestAdd(t *testing.T) {
	t.Parallel()
	ts := fulltext.Tidbits{}
	err := ts.Add("", "")
	be.Err(t, err)
	err = ts.Add("abc", "xyz")
	be.Err(t, err)
}

func TestNewIndex(t *testing.T) {
	t.Parallel()
	ts := fulltext.Tidbits{}
	var fsys embed.FS
	err := ts.NewIndex(fsys, "")
	be.Err(t, err)
}

func TestSearch(t *testing.T) {
	t.Parallel()
	ts := fulltext.Tidbits{}
	r := ts.Search("", 0)
	be.Equal(t, len(r), 0)
}

// go test -fuzz=FuzzSnippet -fuzztime=30s
//

func FuzzSnippet(f *testing.F) {
	// regular string
	f.Add("golang", "The Go programming language is fast.", 5)
	// empty
	f.Add("", "", 0)
	// multibyte unicode
	f.Add("🚀", "Space exploration 🚀 is cool.", 2)

	f.Fuzz(func(t *testing.T, query, body string, window int) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Snippet panicked with query=%q body=%q window=%d: %v", query, body, window, r)
			}
		}()

		result := fulltext.Snippet(query, body, window)
		if query != "" && strings.Contains(strings.ToLower(body), strings.ToLower(query)) {
			if result == "" {
				t.Errorf("Snippet returned empty string for a valid match")
			}
		}
	})
}

// go test -fuzz=FuzzAdd -fuzztime=30s
//

func FuzzAdd(f *testing.F) {
	// Seed with bad content
	f.Add("biography.md", "# John Doe\nThis is **bold** and <script>alert('bad')</script>")
	f.Add("empty.txt", "   ")
	f.Add("legacy.bin", "MacPaint\x00\x9c\xad\xff")

	f.Fuzz(func(t *testing.T, filename, body string) {
		ts := fulltext.Tidbits{}
		ts.New()
		err := ts.Add(filename, body)

		s := strings.TrimSpace(body)
		if filename == "" || s == "" {
			if err == nil {
				t.Errorf("Expected error for empty filename (%q) or body (%q), but got nil", filename, body)
			}
			return
		}
		if err != nil {
			t.Errorf("Unexpected error for filename or body: %v\nfilename: %q\nbody: %q", err, filename, body)
			return
		}

		if ts.Stores() != 1 {
			t.Errorf("Store should have exactly 1 item, got %d\nname: %q\nbody: %q",
				ts.Stores(), filename, body)
		}
		s = ts.Body(0)
		if strings.Contains(s, "<script>") {
			t.Logf("HTML anitization error: %q", s)
		}
		if strings.Contains(s, "**") {
			t.Logf("Markdown sanitization error: %q", s)
		}
		if !utf8.ValidString(s) {
			t.Errorf("Invalid UTF8 in body: %q", s)
		}
	})
}
