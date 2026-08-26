// Package main_test is used for by fuzz tests that need to access the embed data.
package main_test

import (
	"embed"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/tidbit"
)

// NOTE: all fuzz tests must be in root
// go test -fuzz=FuzzSearch -fuzztime=30s
// go test -fuzz=FuzzSearch -fuzztime=10000x

var (
	ts   fulltext.Tidbits //nolint:gochecknoglobals
	once sync.Once        //nolint:gochecknoglobals
	//go:embed public/**/*
	publicFS embed.FS
)

// setupOnce populates the index once for the session.
func setupOnce() {
	once.Do(func() {
		if err := ts.NewIndex(publicFS, tidbit.Dir); err != nil {
			panic(err)
		}
	})
}

func FuzzSearch(f *testing.F) {
	setupOnce()

	f.Add("razor")
	f.Add("defacto2")
	f.Add("razor 1911")
	f.Add("razor 🚀 1911!")
	f.Add("  ")
	f.Add("invalid\x9c\xadbytes")

	f.Fuzz(func(t *testing.T, query string) {
		const avoidExcessRAM = 20
		results := ts.Search(query, avoidExcessRAM)

		if strings.TrimSpace(query) == "" {
			if len(results) != 0 {
				t.Errorf("expected 0 results for empty query, got %d", len(results))
			}
			return
		}

		for _, res := range results {
			if res.Name == "error" {
				t.Errorf("Index Desync: engine returned DocID that doesn't exist in store. Query: %q", query)
			}
			if res.Score <= 0 {
				t.Errorf("Invalid Score: %f for query %q", res.Score, query)
			}
			if res.Name != "" && res.ID <= 0 {
				// Adjust if 0 is a valid ID in your system
				t.Errorf("ID generation failed for Name: %s", res.Name)
			}
		}
	})
}

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
