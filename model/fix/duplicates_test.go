package fix_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/model/fix"
	"github.com/nalgeon/be"
)

func findCluster(t *testing.T, groups []fix.DupGroup, canonical string) fix.DupGroup {
	t.Helper()
	for _, g := range groups {
		if len(g.Entries) > 0 && g.Entries[0].Name == canonical {
			return g
		}
	}
	t.Fatalf("cluster with canonical entry %q not found", canonical)
	return fix.DupGroup{}
}

func TestDuplicates(t *testing.T) {
	t.Parallel()

	rows := []fix.ReleaserCount{
		// Tier 1: Normalized match (numerals, punctuation, stop words)
		{Releaser: "FORUM III BBS", Count: 18},
		{Releaser: "FORUM 3 BBS", Count: 1},
		{Releaser: "PIRATES CLUB", Count: 12},
		{Releaser: "PIRATES CLUB INC", Count: 1},
		// Tier 2: Word swap (higher count listed second to verify canonical sort order)
		{Releaser: "RED SECTOR & TRISTAR", Count: 5},
		{Releaser: "TRISTAR & RED SECTOR", Count: 45},
		// Tier 3: 1-char typo (insertion/deletion, substitution, adjacent transposition)
		{Releaser: "EQUINOX", Count: 25},
		{Releaser: "EQINOX", Count: 1},
		{Releaser: "FAIRLIGHT", Count: 30},
		{Releaser: "FARILIGHT", Count: 2},
		{Releaser: "RAZORBLADE", Count: 15},
		{Releaser: "RAZORBLADS", Count: 1},
		// Non-duplicate numbered boards (must NOT match in Tier 3: sub, transposition, insertion)
		{Releaser: "AREA 51 BBS", Count: 10},
		{Releaser: "AREA 52 BBS", Count: 8},
		{Releaser: "NODE 12 BBS", Count: 6},
		{Releaser: "NODE 21 BBS", Count: 5},
		{Releaser: "TOWER 9 BBS", Count: 7},
		{Releaser: "TOWER 99 BBS", Count: 4},
	}

	rep := fix.AnalyzeDuplicates(rows)
	be.Equal(t, rep.TotalReleasers, len(rows))
	be.Equal(t, rep.TotalClusters, 6)
	be.Equal(t, len(rep.Tier1), 2)
	be.Equal(t, len(rep.Tier2), 1)
	be.Equal(t, len(rep.Tier3), 3)

	forum := findCluster(t, rep.Tier1, "FORUM III BBS")
	be.Equal(t, forum.Kind, "BBS")
	be.Equal(t, forum.Entries[1].Name, "FORUM 3 BBS")

	pirates := findCluster(t, rep.Tier1, "PIRATES CLUB")
	be.Equal(t, pirates.Kind, "Group")
	be.Equal(t, pirates.Entries[1].Name, "PIRATES CLUB INC")

	tristar := findCluster(t, rep.Tier2, "TRISTAR & RED SECTOR")
	be.Equal(t, tristar.Entries[1].Name, "RED SECTOR & TRISTAR")

	eq := findCluster(t, rep.Tier3, "EQUINOX")
	be.Equal(t, eq.Entries[1].Name, "EQINOX")

	flt := findCluster(t, rep.Tier3, "FAIRLIGHT")
	be.Equal(t, flt.Entries[1].Name, "FARILIGHT")

	rb := findCluster(t, rep.Tier3, "RAZORBLADE")
	be.Equal(t, rb.Entries[1].Name, "RAZORBLADS")
}

func TestDuplicatesEdges(t *testing.T) {
	t.Parallel()

	// Empty and single-row inputs
	be.Equal(t, fix.AnalyzeDuplicates(nil).TotalClusters, 0)
	be.Equal(t, fix.AnalyzeDuplicates([]fix.ReleaserCount{{Releaser: "SOLO", Count: 5}}).TotalClusters, 0)
	be.Equal(t, fix.AnalyzeDuplicates([]fix.ReleaserCount{
		{Releaser: "   ", Count: 1},
		{Releaser: "", Count: 2},
	}).TotalClusters, 0)

	// Kind isolation and short-name exclusion (< 6 chars must not match in Tier 3)
	rep := fix.AnalyzeDuplicates([]fix.ReleaserCount{
		{Releaser: "ACME", Count: 10},
		{Releaser: "ACME BBS", Count: 8},
		{Releaser: "ACME FTP", Count: 6},
		{Releaser: "ACID", Count: 20},
		{Releaser: "ACIDS", Count: 1},
	})
	be.Equal(t, rep.TotalClusters, 0)

	// Untrimmed whitespace is flagged in Name while preserving RawName
	ws := fix.AnalyzeDuplicates([]fix.ReleaserCount{
		{Releaser: "ALPHA", Count: 10},
		{Releaser: "ALPHA ", Count: 1},
	})
	be.Equal(t, len(ws.Tier1), 1)
	be.Equal(t, ws.Tier1[0].Entries[0].RawName, "ALPHA")
	be.Equal(t, ws.Tier1[0].Entries[1].RawName, "ALPHA ")
	be.True(t, strings.Contains(ws.Tier1[0].Entries[1].Name, "untrimmed whitespace"))

	// Nil / invalid argument guards
	_, err := fix.Duplicates(t.Context(), nil)
	be.Err(t, err)
	_, err = fix.MergeReleaser(t.Context(), nil, "A", "B")
	be.Err(t, err)
}
