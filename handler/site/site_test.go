package site_test

import (
	"testing"

	"github.com/Defacto2/server/handler/site"
	"github.com/nalgeon/be"
)

func TestFind(t *testing.T) {
	t.Parallel()

	website := site.Find("defacto2")
	if len(website) != 5 {
		t.Fatalf("got %d defacto2 sites, want 5", len(website))
	}
	be.Equal(t, website[0].URL, "https://defacto2.net")
	be.Equal(t, website[0].Name, "Defacto2")
	be.True(t, !website[0].Error)

	be.Equal(t, len(site.Find("notfound")), 0)
	be.Equal(t, len(site.Find("razor-1911-demo")), 2)

	// Working sites (Error: false) must be sorted before defunct sites (Error: true).
	fff := site.Find("fighting-for-fun")
	if len(fff) < 2 {
		t.Fatalf("got %d fighting-for-fun sites, want >= 2", len(fff))
	}
	seenError := false
	for _, s := range fff {
		if s.Error {
			seenError = true
		} else if seenError {
			t.Errorf("working site %q appeared after defunct site", s.URL)
		}
	}
	be.True(t, !fff[0].Error)
	be.True(t, fff[len(fff)-1].Error)
}
