package app_test

import (
	"net/url"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestList(t *testing.T) {
	t.Parallel()

	list := app.List()
	const expectedCount = 9
	be.Equal(t, len(list), expectedCount)
	be.Equal(t, list[0].Name, "Text art scene")

	seenIDs := make(map[string]bool, len(list))
	seenURLs := make(map[string]bool)

	for _, cat := range list {
		be.True(t, cat.Name != "")
		be.True(t, cat.Title != "")
		be.True(t, cat.ID != "")
		be.True(t, cat.ID != "hide")
		be.True(t, !cat.Open)
		be.True(t, len(cat.Sites) > 0)
		be.True(t, !seenIDs[cat.ID])
		seenIDs[cat.ID] = true

		for _, site := range cat.Sites {
			be.True(t, site.Title != "")
			be.True(t, site.Info != "")
			be.True(t, site.URL != "")
			be.True(t, !seenURLs[site.URL])
			seenURLs[site.URL] = true

			parsed, err := url.ParseRequestURI(site.URL)
			be.Err(t, err, nil)
			be.True(t, parsed.Scheme == "http" || parsed.Scheme == "https")
			be.True(t, parsed.Host != "")
		}
	}

	// Mutating Open on one returned accordion must not affect subsequent calls.
	list[0].Open = true
	fresh := app.List()
	be.True(t, !fresh[0].Open)
}
