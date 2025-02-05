package site_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/site"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	t.Parallel()
	website := site.Find("defacto2")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 1)
	assert.Equal(t, "https://defacto2.net", website[0].URL)
	assert.Equal(t, "Defacto2", website[0].Name)
	assert.False(t, website[0].NotWorking)

	website = site.Find("notfound")
	assert.Empty(t, website)

	website = site.Find("razor-1911-demo")
	assert.NotEmpty(t, website)
	assert.Len(t, website, 2)
}

func TestWebsites(t *testing.T) {
	t.Parallel()
	w := site.Websites()
	slices.Sorted(maps.Keys(w))
	for _, key := range w {
		for _, site := range key {
			assert.NotEmpty(t, site.URL)
			assert.NotEmpty(t, site.Name)
			if site.NotWorking {
				// catch any http or https urls
				assert.False(t, strings.HasPrefix(site.URL, "http"),
					"Not working site should not have a http or https URL: %s", site.URL)
				continue
			}
			const localPath = "/"
			if strings.HasPrefix(site.URL, localPath) {
				continue
			}
			assert.True(t, strings.HasPrefix(site.URL, "http"),
				"Working site should have a http or https URL: %s", site.URL)
		}
	}
}
