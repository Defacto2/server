package tidbit_test

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/handler/tidbit"
	"github.com/Defacto2/server/internal/logs"
)

//go:embed testdata/*
var testdata embed.FS

func TestID(t *testing.T) {
	t.Parallel()
	// Test a few known IDs that have URIs defined
	testIDs := []tidbit.ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, id := range testIDs {
		if u := id.URI(); u == nil {
			t.Errorf("tidbit: %d URI is empty", id)
		}
	}
}

func TestMarkdown(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("testdata", "public", "md", "tidbit")
	b := tidbit.ID(1).Markdown(logs.Discard(), testdata, dir)
	if b == nil {
		t.Error("tidbit: 1 markdown is nil")
	}
	const want = "<p>This is a test tidbit.</p>\n"
	if got := string(b); got != want {
		t.Errorf("tidbit: 1 markdown got %q, want %q", got, want)
	}
}

func TestID_URL(t *testing.T) {
	t.Parallel()
	const (
		test1 = "untouchables"
		test2 = "the-untouchables"
		test3 = ""
		want1 = "<a href=\"/g/the-untouchables\">The Untouchables</a>"
		want2 = "<a href=\"/g/untouchables\">Untouchables</a>"
		want3 = "<a href=\"/g/the-untouchables\">The Untouchables</a> &nbsp; <a href=\"/g/untouchables\">Untouchables</a>"
	)
	if got := tidbit.ID(1).URL(test1); got != want1 {
		t.Errorf("tidbit: 1 markdown got %q, want %q", got, want1)
	}
	if got := tidbit.ID(1).URL(test2); got != want2 {
		t.Errorf("tidbit: 1 markdown got %q, want %q", got, want2)
	}
	if got := tidbit.ID(1).URL(test3); got != want3 {
		t.Errorf("tidbit: 1 markdown got %q, want %q", got, want3)
	}
}

func TestFind(t *testing.T) {
	t.Parallel()
	const want = 3
	if got := tidbit.Find("untouchables"); len(got) != want {
		t.Errorf("tidbit: wanted %d untouchables matches, but got %d", want, len(got))
	}
}
