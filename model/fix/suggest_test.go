package fix_test

import (
	"testing"

	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

func TestMatchSuggestions(t *testing.T) {
	t.Parallel()

	rows := []fix.ReleaserCount{
		{Releaser: "RAZOR 1911", Count: 500},
		{Releaser: "ACID PRODUCTIONS", Count: 400},
		{Releaser: "DEFACTO2", Count: 300},
		{Releaser: "EDIE'S BBS", Count: 50},
		{Releaser: "GROUP 2", Count: 40},
		{Releaser: "THE SILENTS", Count: 120},
	}
	entries := fix.BuildSuggestEntries(rows)

	// Tier 1: raw prefix match ordered by count
	got := fix.MatchSuggestions(entries, 10, "razor")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "RAZOR 1911")

	// Tier 2: normalized match (stop words, punctuation, numerals, apostrophes)
	got = fix.MatchSuggestions(entries, 10, "the acid prod")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "ACID PRODUCTIONS")

	got = fix.MatchSuggestions(entries, 10, "edies")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "EDIE'S BBS")

	got = fix.MatchSuggestions(entries, 10, "silents")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "THE SILENTS")

	// Tier 3: 1-char typo match (>= 6 chars)
	got = fix.MatchSuggestions(entries, 10, "defcto2")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "DEFACTO2")

	// Stop-word-only input should not match unrelated entries via normalization
	got = fix.MatchSuggestions(entries, 10, "the ")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "THE SILENTS")

	// Regex metacharacters must not fail
	got = fix.MatchSuggestions(entries, 10, "test[", "group (")
	be.True(t, len(got) >= 0)
}

func TestSuggestReleasersLive(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fix.InvalidateSuggestCache()

	got, err := fix.SuggestReleasers(t.Context(), db, false, 14, "razor")
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)

	// Typo suggestion should still find DEFACTO2
	got, err = fix.SuggestReleasers(t.Context(), db, false, 14, "defcto2")
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)

	// Magazine suggestions
	got, err = fix.SuggestReleasers(t.Context(), db, true, 14, "defacto")
	be.Err(t, err, nil)
	be.True(t, len(got) > 0)
}

func BenchmarkMatchSuggestions(b *testing.B) {
	rows := make([]fix.ReleaserCount, 5000)
	for i := range rows {
		rows[i] = fix.ReleaserCount{Releaser: "SAMPLE GROUP NAME", Count: 5000 - i}
	}
	rows[2500] = fix.ReleaserCount{Releaser: "DEFACTO2", Count: 300}
	entries := fix.BuildSuggestEntries(rows)

	b.ResetTimer()
	for range b.N {
		_ = fix.MatchSuggestions(entries, 14, "defcto2")
	}
}
