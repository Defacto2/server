// Package main_test is used for by fuzz tests that need to access the embed data.
package main_test

import (
	"embed"
	"strings"
	"sync"
	"testing"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/handler/tidbit"
)

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

// go test -fuzz=FuzzSearch -fuzztime=30s
//

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
