//nolint:staticcheck,wastedassign
package areacode_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/areacode"
	"github.com/nalgeon/be"
)

func TestTerritories(t *testing.T) {
	t.Parallel()

	ca := areacode.CanadaProvinces()
	us := areacode.USAStates()
	misc := areacode.Miscellaneous()
	be.Equal(t, len(ca), 13)
	be.Equal(t, len(us), 51)
	be.Equal(t, len(misc), 2)

	regs := areacode.Regions()
	be.Equal(t, len(regs), len(ca)+len(us)+len(misc))
	be.True(t, slices.IsSortedFunc(regs, func(a, b areacode.Region) int {
		return strings.Compare(a.Name, b.Name)
	}))

	abbrs := areacode.Abbreviations()
	be.True(t, len(abbrs) > 0)
	be.True(t, slices.IsSorted(abbrs))

	codes := areacode.AreaCodes()
	be.True(t, len(codes) > 0)
	be.True(t, slices.IsSorted(codes))

	for nan := range areacode.All() {
		be.True(t, nan.Valid())
		be.True(t, len(areacode.RegionByCode(nan)) >= 1)
	}
}

func TestHTML(t *testing.T) {
	t.Parallel()

	t.Run("NAN", func(t *testing.T) {
		t.Parallel()

		be.Equal(t, string(areacode.NAN(0).HTML()), "")
		be.Equal(t, string(areacode.NAN(999).HTML()), "")
		be.Equal(t, string(areacode.NAN(602).HTML()), "<span>602 - Arizona (AZ)</span><br>")
		be.Equal(t, string(areacode.NAN(403).HTML()), "<span>403 - Alberta (AB) + Yukon (YT)</span><br>")

		withNote := string(areacode.NAN(201).HTML())
		be.True(t, strings.HasPrefix(withNote, "<span>201 - New Jersey (NJ) <small><em>"))
		be.True(t, strings.HasSuffix(withNote, "</em></small></span><br>"))

		noAbbr := string(areacode.NAN(710).HTML())
		be.True(t, strings.HasPrefix(noAbbr, "<span>710 - United States Government <small><em>"))
	})

	t.Run("Abbreviation", func(t *testing.T) {
		t.Parallel()

		be.Equal(t, string(areacode.Abbreviation("ZZ").HTML()), "<span>ZZ - n/a</span><br>")
		be.Equal(t, string(areacode.Abbreviation("AB").HTML()), "<span>AB (Alberta) - 403</span><br>")
		be.Equal(t, string(areacode.Abbreviation("CO").HTML()), "<span>CO (Colorado) - 303, 719</span><br>")
	})

	t.Run("Region", func(t *testing.T) {
		t.Parallel()

		nunavut := areacode.RegionByName("Nunavut")
		be.Equal(t, string(nunavut.HTML()), "<span>Nunavut (NU) - n/a</span><br>")

		colorado := areacode.RegionByName("Colorado")
		be.Equal(t, string(colorado.HTML()), "<span>Colorado (CO) - 303, 719</span><br>")

		gov := areacode.RegionByName("United States Government")
		be.Equal(t, string(gov.HTML()), "<span>United States Government - 710</span><br>")
	})
}

func TestLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  int
	}{
		{"empty string", "", 0},
		{"single char", "a", 0},
		{"unknown abbr", "ZZ", 0},
		{"valid abbr", "tx", 1},
		{"substring match", "carolina", 2},
		{"substring miss", "atlantis", 0},
		{"negative int", -1, 0},
		{"large int", 1000, 0},
		{"valid int", 212, 1},
		{"large uint", uint(1000), 0},
		{"valid uint", uint(212), 1},
		{"valid NAN", areacode.NAN(902), 2},
		{"invalid NAN", areacode.NAN(999), 0},
		{"unsupported type", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := areacode.Lookup(tt.input)
			be.Equal(t, len(got), tt.want)
		})
	}
}

func TestLookups(t *testing.T) {
	t.Parallel()

	// Deduplicates repeated matches and skips unmatched inputs.
	got := areacode.Lookups("tx", 214, "texas", "unknown", 999, true)
	be.Equal(t, len(got), 1)
	be.Equal(t, got[0].Abbreviation, areacode.Abbreviation("TX"))
}

func TestQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		wantCode areacode.NAN
		wantRegs int
	}{
		{"string code", "602", 602, 0},
		{"string abbr", "tx", 0, 1},
		{"string miss", "atlantis", 0, 0},
		{"int", 212, 212, 0},
		{"int negative", -1, 0, 0},
		{"int too large", 1000, 0, 0},
		{"uint valid", uint(212), 212, 0},
		{"uint too large", uint(1000), 0, 0},
		{"NAN", areacode.NAN(415), 415, 0},
		{"unsupported", true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := areacode.Query(tt.input)
			be.Equal(t, got.AreaCode, tt.wantCode)
			be.Equal(t, len(got.Region), tt.wantRegs)
		})
	}
}

func TestQueries(t *testing.T) {
	t.Parallel()

	be.Equal(t, len(areacode.Queries(" CA ", "", "  ", "201", "999", "atlantis")), 2)

	// Exceeding the maximum cap stops collecting further queries.
	many := make([]string, 120)
	for i := range many {
		many[i] = "212"
	}
	be.Equal(t, len(areacode.Queries(many...)), 100)
}

func TestRegionLookups(t *testing.T) {
	t.Parallel()

	_, ok := areacode.RegionByAbbr("")
	be.Equal(t, ok, false)
	_, ok = areacode.RegionByAbbr("ZZ")
	be.Equal(t, ok, false)
	ct, ok := areacode.RegionByAbbr("ct")
	be.Equal(t, ok, true)
	be.Equal(t, ct.Name, "Connecticut")

	be.Equal(t, len(areacode.RegionByCode(0)), 0)
	be.Equal(t, len(areacode.RegionByCode(199)), 0)
	be.Equal(t, len(areacode.RegionByCode(1000)), 0)
	be.Equal(t, len(areacode.RegionByCode(902)), 2)

	be.Equal(t, areacode.RegionByName("ontario").Abbreviation, areacode.Abbreviation("ON"))
	be.Equal(t, areacode.RegionByName("nonexistent").Name, "")

	be.True(t, areacode.RegionContains("") == nil)
	be.Equal(t, len(areacode.RegionContains("south")), 2)
	be.True(t, areacode.RegionContains("nonexistent") == nil)

	be.Equal(t, areacode.Contains(ct, areacode.Regions()), true)
	be.Equal(t, areacode.Contains(areacode.Region{Name: "Atlantis"}, areacode.Regions()), false)
}
