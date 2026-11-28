package htmx

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestPingHelpers(t *testing.T) {
	t.Parallel()

	be.Equal(t, pingStatusClass(http.StatusOK), "text-success")
	be.Equal(t, pingStatusClass(http.StatusNotFound), "text-danger-emphasis")
	be.Equal(t, pingStatusClass(http.StatusInternalServerError), "text-danger")
	be.Equal(t, pingStatusClass(http.StatusServiceUnavailable), "text-danger")
	be.Equal(t, pingStatusClass(http.StatusFound), "text-warning-emphasis")

	// Closed port formats error div with escaped path.
	got := pingResultHTML(t.Context(), "/bad<path>", "http", 1)
	be.True(t, strings.Contains(got, "/bad&lt;path&gt;"))
}

func TestTemplateHelpers(t *testing.T) {
	t.Parallel()

	be.Equal(t, emptyFS(nil), true)
	be.Equal(t, emptyFS(embed.FS{}), true)
	be.Equal(t, emptyFS(testutil.ProjectFS(t)), false)

	// borderClass: exact match, initialism match, and default.
	be.Equal(t, borderClass("razor", "razor"), "border border-primary")
	init := initialisms("razor-1911")
	be.True(t, init != "")
	be.Equal(t, borderClass(init, "razor-1911"), "border border-primary")
	be.Equal(t, borderClass("other", "razor-1911"), "border")

	// fmtPath: known mapped path vs fallback humanization.
	be.True(t, fmtPath("razor-1911") != "")
	be.Equal(t, fmtPath("some-unknown-group"), "Some Unknown Group")

	// state: removed, not approved, and online.
	be.Equal(t, state(false, false), template.HTML(`<span title="Removed from public">🚫</span>`))
	be.Equal(t, state(false, true), template.HTML(`<span title="Not approved">⛔</span>`))
	be.Equal(t, state(true, false), template.HTML(""))
	be.Equal(t, state(true, true), template.HTML(""))

	// suggestHint: empty, value match, label match, and fuzzy mismatch.
	be.Equal(t, suggestHint("Razor 1911", "Razor 1911 (5 items)", ""), "")
	be.Equal(t, suggestHint("Razor 1911", "Razor 1911 (5 items)", "   "), "")
	be.Equal(t, suggestHint("Razor 1911", "Razor 1911 (5 items)", "razor"), "")
	be.Equal(t, suggestHint("TRSi", "Tristar & Red Sector Inc.", "tristar"), "")
	be.Equal(t, suggestHint("Razor 1911", "Razor 1911 (5 items)", "rzr"), " ≈ rzr")
}
