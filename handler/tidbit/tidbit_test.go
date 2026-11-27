package tidbit_test

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestID(t *testing.T) {
	t.Parallel()

	for _, id := range []tidbit.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		be.True(t, len(id.URI()) > 0)
	}
	be.True(t, tidbit.ID(-1).URI() == nil)
	be.True(t, tidbit.ID(999999).URI() == nil)
}

func TestMarkdown(t *testing.T) {
	t.Parallel()

	fsys := testutil.EmbedFS(t)
	const dir = "testdata/tidbit"

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		b := tidbit.ID(1).Markdown(testLog, fsys, dir)
		be.Equal(t, string(b), "<p>This is a test tidbit.</p>\n")
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&buf, nil))
		b := tidbit.ID(999999).Markdown(sl, fsys, dir)
		be.True(t, b == nil)
		be.True(t, buf.Len() > 0)
	})
}

func TestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   tidbit.ID
		uri  string
		want template.HTML
	}{
		{"exclude untouchables", 1, "untouchables", `<a href="/g/the-untouchables">The Untouchables</a>`},
		{"exclude the-untouchables", 1, "the-untouchables", `<a href="/g/untouchables">Untouchables</a>`},
		{"exclude none", 1, "", `<a href="/g/the-untouchables">The Untouchables</a> &nbsp; <a href="/g/untouchables">Untouchables</a>`},
		{"negative id", -1, "", ""},
		{"unknown id", 999999, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.id.URL(tt.uri), tt.want)
		})
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(tidbit.Find("untouchables")), 3)
	be.Equal(t, len(tidbit.Find("no-such-group")), 0)
}

func TestMatch(t *testing.T) {
	t.Parallel()

	be.True(t, tidbit.Missing("no-such-group"))

	for _, uris := range tidbit.Copy() {
		for _, uri := range uris {
			key := string(uri)
			be.True(t, !tidbit.Missing(key))
			be.True(t, len(tidbit.Find(key)) > 0)
		}
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)
	got := tidbit.ID(1).String(testLog, fsys)
	be.True(t, strings.Contains(got, "<p>"))

	be.Equal(t, tidbit.ID(999999).String(testLog, fsys), "")
}

func TestFiles(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)
	entries, err := fs.ReadDir(fsys, tidbit.Dir)
	be.Err(t, err, nil)

	files := make(map[tidbit.ID]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stem, ok := strings.CutSuffix(entry.Name(), ".md")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(stem)
		be.Err(t, err, nil)
		files[tidbit.ID(n)] = true
	}

	index := tidbit.Copy()
	for id := range index {
		if !files[id] {
			t.Errorf("tidbit ID %d has no markdown file in %s", id, tidbit.Dir)
		}
	}
	for id := range files {
		if _, ok := index[id]; !ok {
			t.Errorf("markdown file %s/%d.md is not registered in tidbit index", tidbit.Dir, id)
		}
	}
}

func TestStartup(t *testing.T) {
	t.Parallel()

	tidbit.Startup(t.Context(), testLog, nil)
}
