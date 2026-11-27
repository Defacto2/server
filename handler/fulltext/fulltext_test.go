package fulltext_test

import (
	"embed"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	t.Run("no index", func(t *testing.T) {
		t.Parallel()

		ts := fulltext.Tidbits{}
		be.Err(t, ts.Add("", ""), fulltext.ErrNoIndex)
		be.Err(t, ts.Add("abc", "xyz"), fulltext.ErrNoIndex)
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()

		ts := fulltext.Tidbits{}
		ts.New()
		be.Err(t, ts.Add("", "body"), fulltext.ErrNoName)
		be.Err(t, ts.Add("1.md", ""), fulltext.ErrNoBody)
		be.Err(t, ts.Add("1.md", "  \t\n "), fulltext.ErrNoBody)
	})

	t.Run("strips html markdown and box chars", func(t *testing.T) {
		t.Parallel()

		ts := fulltext.Tidbits{}
		ts.New()
		err := ts.Add("42.md", `<a href="https://example.com">**Razor 1911**</a> ┌──┐ demo`)
		be.Err(t, err, nil)
		be.Equal(t, ts.Stores(), 1)
		be.Equal(t, ts.Name(0), "42.md")
		body := ts.Body(0)
		be.True(t, strings.Contains(body, "Razor 1911"))
		be.True(t, !strings.Contains(body, "<a"))
		be.True(t, !strings.Contains(body, "**"))
		be.True(t, !strings.Contains(body, "┌"))

		docs, terms := ts.Stats()
		be.Equal(t, docs, 1)
		be.True(t, terms > 0)

		ts.New()
		be.Equal(t, ts.Stores(), 0)
		docs, terms = ts.Stats()
		be.Equal(t, docs, 0)
		be.Equal(t, terms, int64(0))
	})
}

func TestAccessors(t *testing.T) {
	t.Parallel()

	ts := fulltext.Tidbits{}
	be.Equal(t, ts.Stores(), 0)
	be.Equal(t, ts.Body(0), "")
	be.Equal(t, ts.Name(0), "error")

	ts.New()
	be.Equal(t, ts.Body(-1), "")
	be.Equal(t, ts.Body(0), "")
	be.Equal(t, ts.Name(-1), "error")
	be.Equal(t, ts.Name(0), "error")
}

//nolint:paralleltest // NewIndex temporarily replaces global slog.Default().
func TestNewIndex(t *testing.T) {
	t.Run("invalid fs and root", func(t *testing.T) {
		ts := fulltext.Tidbits{}
		var efs embed.FS
		be.Err(t, ts.NewIndex(efs, ""))

		fsys := testutil.ProjectFS(t)
		be.Err(t, ts.NewIndex(fsys, "/"))
	})

	t.Run("indexes files and skips empty", func(t *testing.T) {
		fsys := fstest.MapFS{
			"tidbit/118.md":     {Data: []byte("**ACRONYM** was a short-lived PC cracking group.")},
			"tidbit/sub/343.md": {Data: []byte("CORE released DOS tools and demos.")},
			"tidbit/empty.md":   {Data: []byte("   \n\t  ")},
		}

		ts := fulltext.Tidbits{}
		be.Err(t, ts.NewIndex(fsys, "tidbit"), nil)
		be.Equal(t, ts.Stores(), 2)

		docs, terms := ts.Stats()
		be.Equal(t, docs, 2)
		be.True(t, terms > 0)

		// Re-indexing overwrites the previous index.
		fsys2 := fstest.MapFS{
			"tidbit/99.md": {Data: []byte("Defacto2 scene archive.")},
		}
		be.Err(t, ts.NewIndex(fsys2, "tidbit"), nil)
		be.Equal(t, ts.Stores(), 1)
		docs, _ = ts.Stats()
		be.Equal(t, docs, 1)
		be.Equal(t, ts.Name(0), "99.md")
	})
}

//nolint:paralleltest // NewIndex temporarily replaces global slog.Default().
func TestSearch(t *testing.T) {
	empty := fulltext.Tidbits{}
	be.Equal(t, len(empty.Search("razor", 10)), 0)

	fsys := fstest.MapFS{
		"docs/118.md":   {Data: []byte("Razor 1911 released classic DOS demos and cracks.")},
		"docs/200.md":   {Data: []byte("Future Crew created Second Reality demo.")},
		"docs/notes.md": {Data: []byte("Non-numeric markdown file about Amiga crackers.")},
		"docs/info.txt": {Data: []byte("Plain text file mentioning Fairlight.")},
	}

	ts := fulltext.Tidbits{}
	be.Err(t, ts.NewIndex(fsys, "docs"), nil)

	t.Run("empty query", func(t *testing.T) {
		be.Equal(t, len(ts.Search("", 10)), 0)
		be.Equal(t, len(ts.Search("   ", 10)), 0)
		be.Equal(t, len(ts.Search(",,, ;;;", 10)), 0)
	})

	t.Run("single term", func(t *testing.T) {
		got := ts.Search("Razor", 10)
		be.Equal(t, len(got), 1)
		be.Equal(t, got[0].ID, 118)
		be.Equal(t, got[0].Name, "118.md")
		be.True(t, got[0].Score > 0)
		be.True(t, strings.Contains(got[0].Snip, "Razor"))
	})

	t.Run("multi term and", func(t *testing.T) {
		for _, q := range []string{"Future Reality", "Future,Reality", "Future;Reality"} {
			got := ts.Search(q, 10)
			be.Equal(t, len(got), 1)
			be.Equal(t, got[0].ID, 200)
		}
	})

	t.Run("no match", func(t *testing.T) {
		be.Equal(t, len(ts.Search("nonexistentterm", 10)), 0)
	})

	t.Run("invalid id filenames", func(t *testing.T) {
		got := ts.Search("Amiga", 10)
		be.Equal(t, len(got), 1)
		be.Equal(t, got[0].ID, -1)

		got = ts.Search("Fairlight", 10)
		be.Equal(t, len(got), 1)
		be.Equal(t, got[0].ID, -1)
	})
}

func TestSnippet(t *testing.T) {
	t.Parallel()

	body := "one two three four five six seven eight nine ten"
	tests := []struct {
		name   string
		query  string
		body   string
		window int
		want   string
	}{
		{"empty query", "", body, 2, ""},
		{"negative window", "five", body, -5, "...five..."},
		{"match at start", "one", body, 2, "one two..."},
		{"match in middle", "five", body, 2, "...four five six..."},
		{"match at end", "ten", body, 2, "...nine ten"},
		{"case insensitive", "FIVE", "four five six", 5, "four five six"},
		{"no match truncates", "missing", body, 2, "one two three four..."},
		{"no match short body", "missing", "short text", 5, "short text..."},
		{"utf8 match", "CAFÉ", "visit the café today", 2, "...the café today"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fulltext.Snippet(tt.query, tt.body, tt.window)
			be.Equal(t, got, tt.want)
		})
	}
}
