package site_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/site"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()
	website := site.Find("defacto2")
	be.True(t, len(website) == 5)
	be.Equal(t, "https://defacto2.net", website[0].URL)
	be.Equal(t, "Defacto2", website[0].Name)
	be.True(t, !website[0].NotWorking)

	website = site.Find("notfound")
	be.True(t, len(website) == 0)
	website = site.Find("razor-1911-demo")
	be.True(t, len(website) == 2)
}

func TestWebsites(t *testing.T) {
	t.Parallel()
	w := site.Websites()
	for _, key := range w {
		for _, site := range key {
			be.True(t, site.URL != "")
			be.True(t, site.Name != "")
			if site.NotWorking {
				// catch any http or https urls
				be.True(t, !strings.HasPrefix(site.URL, "http"))
				continue
			}
			const localPath = "/"
			if strings.HasPrefix(site.URL, localPath) {
				continue
			}
			be.True(t, strings.HasPrefix(site.URL, "http"))
		}
	}
}
