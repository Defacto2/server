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
		{Releaser: "FORUM 3 BBS", Count: 40},
		{Releaser: "THE SILENTS", Count: 120},
		{Releaser: "PSYKLE FTP", Count: 30},
		{Releaser: "LORDS OF DARKNESS", Count: 25},
	}
	entries := fix.Suggest(rows)

	// Tier 1: raw prefix match ordered by count
	got := fix.Suggests(entries, 10, "razor")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "RAZOR 1911")

	// Tier 2: normalized match (stop words, punctuation, numerals, apostrophes)
	got = fix.Suggests(entries, 10, "the acid prod")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "ACID PRODUCTIONS")

	got = fix.Suggests(entries, 10, "edies")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "EDIE'S BBS")

	got = fix.Suggests(entries, 10, "silents")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "THE SILENTS")

	got = fix.Suggests(entries, 10, "forum iii")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "FORUM 3 BBS")

	// Tier 2b: word-order swap match
	got = fix.Suggests(entries, 10, "Darkness of Lords")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "LORDS OF DARKNESS")

	// Tier 3: 1-char typo match (>= 6 chars)
	got = fix.Suggests(entries, 10, "defcto2")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "DEFACTO2")

	got = fix.Suggests(entries, 10, "Psycle ftp")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "PSYKLE FTP")

	// Stop-word-only input should not match unrelated entries via normalization
	got = fix.Suggests(entries, 10, "the ")
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Releaser, "THE SILENTS")

	// Regex metacharacters must not panic or fail
	got = fix.Suggests(entries, 10, "test[", "group (")
	be.Equal(t, len(got), 0)

	// Guard cases
	be.Equal(t, len(fix.Suggests(nil, 10, "razor")), 0)
	be.Equal(t, len(fix.Suggests(entries, 0, "razor")), 0)
	be.Equal(t, len(fix.Suggests(entries, 10, "   ")), 0)

	_, err := fix.SuggestReleasers(t.Context(), nil, false, 10, "razor")
	be.Err(t, err)
}

func TestSuggestsPriorityAndLimit(t *testing.T) {
	t.Parallel()

	entries := fix.Suggest([]fix.ReleaserCount{
		{Releaser: "DEFACTO2", Count: 50},     // Tier 1 (raw prefix)
		{Releaser: "DEFACTO", Count: 100},     // Tier 1 (higher count)
		{Releaser: "THE DEFACTO", Count: 500}, // Tier 2 (normalized match, higher count than Tier 1)
		{Releaser: "DEFACXO", Count: 900},     // Tier 3 (1-char typo, highest count overall)
		{Releaser: "UNRELATED", Count: 999},   // No match
	})

	// Tier 1 (ordered by count) precedes Tier 2, which precedes Tier 3
	got := fix.Suggests(entries, 10, "defacto")
	be.Equal(t, len(got), 4)
	be.Equal(t, got[0].Releaser, "DEFACTO")
	be.Equal(t, got[1].Releaser, "DEFACTO2")
	be.Equal(t, got[2].Releaser, "THE DEFACTO")
	be.Equal(t, got[3].Releaser, "DEFACXO")

	// Limit truncates cleanly across tiers
	got = fix.Suggests(entries, 3, "defacto")
	be.Equal(t, len(got), 3)
	be.Equal(t, got[0].Releaser, "DEFACTO")
	be.Equal(t, got[1].Releaser, "DEFACTO2")
	be.Equal(t, got[2].Releaser, "THE DEFACTO")
}

func TestSuggestReleasersLive(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	fix.Clear()

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
	entries := fix.Suggest(rows)

	for b.Loop() {
		_ = fix.Suggests(entries, 14, "defcto2")
	}
}
