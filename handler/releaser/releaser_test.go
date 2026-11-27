package releaser_test

import (
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/nalgeon/be"
)

func TestCell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty string", "", ""},
		{"single word", "defacto2.net", "DEFACTO2NET"},
		{"leading the", "the blah", "THE BLAH"},
		{"common the", "in the blah", "IN THE BLAH"},
		{"no spaces", "TheBlah", "THEBLAH"},
		{"elite fmt", "MiRROR now", "MIRROR NOW"},
		{"roman numbers", "In the row now ii", "IN THE ROW NOW II"},
		{"BBS", "MiRROR now bbS", "MIRROR NOW BBS"},
		{"slug", "this-is-a-slug-string", "THIS-IS-A-SLUG-STRING"},
		{
			"pair of groups",
			"Group inc.,RAZOR TO 1911",
			"GROUP INC, RAZOR TO 1911",
		},
		{
			"2nd group with a leading the",
			"this is the group,the group is this",
			"THIS IS THE GROUP, THE GROUP IS THIS",
		},
		{"ordinal", "4TH dimension", "4TH DIMENSION"},
		{"ordinals", "4TH dimension, 5Th Dynasty", "4TH DIMENSION, 5TH DYNASTY"},
		{"abbreviation", "2000 ad", "2000 AD"},
		{
			"mega-group",
			"Lightforce,Pact,TRSi,Venom,Razor 1911,the System",
			"LIGHTFORCE, PACT, TRSI, VENOM, RAZOR 1911, THE SYSTEM",
		},
		{"coop", "coop", "COOP"},
		{"example 1", "the  Defacto2  demo  group", "THE DEFACTO2 DEMO GROUP"},
		{"example 2", "  the x bbs  ", "X BBS"},
		{"example 3", "TDT / TRSi", "TDT TRSI"},
		{"example 4", "TDT,TRSi", "TDT, TRSI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Cell(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestClean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty string", "", ""},
		{"leading the", "the blah", "The Blah"},
		{"common the", "in the blah", "In the Blah"},
		{"no spaces", "TheBlah", "Theblah"},
		{"elite fmt", "MiRROR now", "Mirror Now"},
		{"roman numbers", "In the row now ii", "In the Row Now II"},
		{"BBS", "MiRROR now bbS", "Mirror Now BBS"},
		{"slug", "this-is-a-slug-string", "This-is-a-Slug-String"},
		{
			"pair of groups",
			"Group inc.,RAZOR TO 1911",
			"Group Inc, Razor to 1911",
		},
		{
			"2nd group with a leading the",
			"this is the group,the group is this",
			"This is the Group, The Group is This",
		},
		{"ordinal", "4TH dimension", "4th Dimension"},
		{"ordinals", "4TH dimension, 5Th Dynasty", "4th Dimension, 5th Dynasty"},
		{"abbreviation", "2000 ad", "2000AD"},
		{"abbreviations", "2000ad, 500bc", "2000AD, 500BC"},
		{
			"mega-group",
			"Lightforce,Pact,TRSi,Venom,Razor 1911,the System",
			"Lightforce, Pact, TRSi, Venom, Razor 1911, The System",
		},
		{"example 1", "the  Defacto2  demo  group", "The Defacto2 Demo Group"},
		{"example 2", "  the x bbs  ", "X BBS"},
		{"example 3", "The X Ftp", "X FTP"},
		{"example 4", "tdt / trsi", "Tdt Trsi"},
		{"example 5", "tdt,trsi", "Tdt, TRSi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Clean(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestHumanize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty", "", ""},
		{"single", "defacto2", "Defacto2"},
		{"invalid", "/razor-1911//", ""},
		{"ampersand", "razor-1911-ampersand-skillion", "Razor 1911 & Skillion"},
		{"pair", "razor-1911*trsi", "Razor 1911, TRSi"},
		{"underscore", "north-american-pirate_phreak-association", "North American Pirate-Phreak Association"},
		{"bbs", "2-minutes-to-midnight-bbs", "2 Minutes to Midnight BBS"},
		{"abbreviation", "2000ad", "2000AD"},
		{"leading digit", "2tally-unrubbed", "2Tally Unrubbed"},
		{"alphanumeric bbs", "2nd2none-bbs", "2ND2NONE BBS"},
		{"multi", "class*paradigm*razor-1911", "Class, Paradigm, Razor 1911"},
		{"multi bbs", "down-town-bbs*bizare-bbs", "Down Town BBS, Bizare BBS"},
		{"special collab", "united-software-association*fairlight", "United Software Association + Fairlight PC Division"},
		{"special coop", "coop", "TDT / TRSi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Humanize(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty", "", ""},
		{"invalid", "/razor-1911//", ""},
		{"single", "defacto2", "DEFACTO2"},
		{"uppercase", "Razor-1911", "RAZOR 1911"},
		{"ampersand", "razor-1911-ampersand-skillion", "RAZOR 1911 & SKILLION"},
		{"underscore", "north-american-pirate_phreak-association", "NORTH AMERICAN PIRATE-PHREAK ASSOCIATION"},
		{"multi", "class*paradigm*razor-1911", "CLASS, PARADIGM, RAZOR 1911"},
		{"special", "coop", "COOP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Index(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty", "", ""},
		{"invalid", "/invalid/path", ""},
		{"single", "class", "Class"},
		{"multi", "class*paradigm*razor-1911", "Class + Paradigm + Razor 1911"},
		{"special collab", "united-software-association*fairlight", "United Software Association + Fairlight PC Division"},
		{"special coop", "coop", "TDT / TRSi"},
		{"pair", "razor-1911-demo*trsi", "Razor 1911 Demo + TRSi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Link(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestObfuscate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty string", "", ""},
		{"single word", "hello", "hello"},
		{"coop", "coop", "coop"},
		{"tdt/trsi", "tdt / trsi", "coop"},
		{"multiple words", "the quick brown fox", "the-quick-brown-fox"},
		{"special characters", "h3ll0 w0rld!", "h3ll0-w0rld"},
		{"ampersand and comma", "hello & world, foxes", "hello-ampersand-world*foxes"},
		{"initialism", "nappa", "north-american-pirate_phreak-association"},
		{"readme example 1", "The 12AM BBS.", "12am-bbs"},
		{"readme example 2", "ACiD Productions", "acid-productions"},
		{"readme example 3", "Razor 1911 Demo & Skillion", "razor-1911-demo-ampersand-skillion"},
		{"readme example 4", "TDU-Jam!", "tdu_jam"},
		{
			"readme example 5", "United Software Association + Fairlight PC Division",
			"united-software-association*fairlight",
		},
		{"readme example 6", "TDT", "the-dream-team"},
		{"readme example 7", "fltdox", "fairlight-dox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Obfuscate(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"empty string", "", ""},
		{"standard", "razor 1911", "Razor 1911"},
		{"casing", " _.=[   RaZoR 1911   ]=._ ", "Razor 1911"},
		{"special slug", "coop", "TDT / TRSi"},
		{"special name", "tdt / trsi", "TDT / TRSi"},
		{"initialism", "nappa", "North American Pirate-Phreak Association"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Title(tt.arg)
			be.Equal(t, got, tt.want)
		})
	}
}
