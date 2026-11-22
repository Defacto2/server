package fix_test

import (
	"testing"

	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

func TestAnalyzeDuplicates(t *testing.T) {
	t.Parallel()

	rows := []fix.ReleaserCount{
		// Tier 1: Normalized match (numerals, punctuation, stop words)
		{Releaser: "FORUM III BBS", Count: 18},
		{Releaser: "FORUM 3 BBS", Count: 1},
		{Releaser: "PIRATES CLUB", Count: 12},
		{Releaser: "PIRATES CLUB INC", Count: 1},
		// Tier 2: Word swap
		{Releaser: "RED SECTOR & TRISTAR", Count: 5},
		{Releaser: "TRISTAR & RED SECTOR", Count: 45},
		// Tier 3: 1-char typo / transposition
		{Releaser: "EQUINOX", Count: 25},
		{Releaser: "EQINOX", Count: 1},
		// Non-duplicate numbered boards (must NOT match in Tier 3)
		{Releaser: "AREA 51 BBS", Count: 10},
		{Releaser: "AREA 52 BBS", Count: 8},
	}

	rep := fix.AnalyzeDuplicates(rows)
	be.Equal(t, rep.TotalReleasers, len(rows))
	be.Equal(t, len(rep.Tier1), 2)
	be.Equal(t, len(rep.Tier2), 1)
	be.Equal(t, len(rep.Tier3), 1)
	be.Equal(t, rep.Tier3[0].Entries[0].Name, "EQUINOX")
	be.Equal(t, rep.Tier3[0].Entries[1].Name, "EQINOX")
}
