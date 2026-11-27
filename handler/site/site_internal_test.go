package site

import (
	"strings"
	"testing"
)

func TestWebsites(t *testing.T) {
	t.Parallel()

	for uri, sites := range websites {
		seenError := false
		for _, s := range sites {
			if s.URL == "" {
				t.Errorf("%s: empty URL", uri)
			}
			if s.URL != strings.TrimSpace(s.URL) {
				t.Errorf("%s: URL %q has leading or trailing whitespace", uri, s.URL)
			}
			if !s.Error && !strings.HasPrefix(s.URL, "http://") && !strings.HasPrefix(s.URL, "https://") {
				t.Errorf("%s: working site URL %q must include http:// or https://", uri, s.URL)
			}
			if s.Error {
				seenError = true
			} else if seenError {
				t.Errorf("%s: working site %q appeared after defunct site", uri, s.URL)
			}
		}
	}
}
