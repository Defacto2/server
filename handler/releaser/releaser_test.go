package releaser_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/nalgeon/be"
)

func TestCell(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty string", args{""}, ""},
		{"single word", args{"defacto2.net"}, "defacto2net"},
		{"leading the", args{"the blah"}, "The Blah"},
		{"common the", args{"in the blah"}, "In the Blah"},
		{"no spaces", args{"TheBlah"}, "Theblah"},
		{"elite fmt", args{"MiRROR now"}, "Mirror Now"},
		{"roman numbers", args{"In the row now ii"}, "In the Row Now II"},
		{"BBS", args{"MiRROR now bbS"}, "Mirror Now BBS"},
		{"slug", args{"this-is-a-slug-string"}, "This-is-a-Slug-String"},
		{
			"pair of groups",
			args{"Group inc.,RAZOR TO 1911"},
			"Group Inc, Razor to 1911",
		},
		{
			"2nd group with a leading the",
			args{"this is the group,the group is this"},
			"This is the Group, The Group is This",
		},
		{"ordinal", args{"4TH dimension"}, "4th Dimension"},
		{"ordinals", args{"4TH dimension, 5Th Dynasty"}, "4th Dimension, 5th Dynasty"},
		{"abbreviation", args{"2000 ad"}, "2000 AD"},
		{
			"mega-group",
			args{"Lightforce,Pact,TRSi,Venom,Razor 1911,the System"},
			"Lightforce, Pact, Trsi, Venom, Razor 1911, The System",
		},
		{"coop", args{"coop"}, "COOP"},
		{"example 1", args{"the  Defacto2  demo  group"}, "The Defacto2 Demo Group"},
		{"example 2", args{"  the x bbs  "}, "X BBS"},
		{"example 3", args{"TDT / TRSi"}, "TDT TRSI"},
		{"example 4", args{"TDT,TRSi"}, "TDT, TRSI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := strings.ToUpper(tt.want)
			got := releaser.Cell(tt.args.s)
			be.Equal(t, got, want)
		})
	}
}

func TestClean(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty string", args{""}, ""},
		{"leading the", args{"the blah"}, "The Blah"},
		{"common the", args{"in the blah"}, "In the Blah"},
		{"no spaces", args{"TheBlah"}, "Theblah"},
		{"elite fmt", args{"MiRROR now"}, "Mirror Now"},
		{"roman numbers", args{"In the row now ii"}, "In the Row Now II"},
		{"BBS", args{"MiRROR now bbS"}, "Mirror Now BBS"},
		{"slug", args{"this-is-a-slug-string"}, "This-is-a-Slug-String"},
		{
			"pair of groups",
			args{"Group inc.,RAZOR TO 1911"},
			"Group Inc, Razor to 1911",
		},
		{
			"2nd group with a leading the",
			args{"this is the group,the group is this"},
			"This is the Group, The Group is This",
		},
		{"ordinal", args{"4TH dimension"}, "4th Dimension"},
		{"ordinals", args{"4TH dimension, 5Th Dynasty"}, "4th Dimension, 5th Dynasty"},
		{"abbreviation", args{"2000 ad"}, "2000AD"},
		{"abbreviations", args{"2000ad, 500bc"}, "2000AD, 500BC"},
		{
			"mega-group",
			args{"Lightforce,Pact,TRSi,Venom,Razor 1911,the System"},
			"Lightforce, Pact, TRSi, Venom, Razor 1911, The System",
		},
		{"example 1", args{"the  Defacto2  demo  group"}, "The Defacto2 Demo Group"},
		{"example 2", args{"  the x bbs  "}, "X BBS"},
		{"example 3", args{"The X Ftp"}, "X FTP"},
		{"example 4", args{"tdt / trsi"}, "Tdt Trsi"},
		{"example 5", args{"tdt,trsi"}, "Tdt, TRSi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := releaser.Clean(tt.args.s)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestHumanize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "defacto2",
			want:  "Defacto2",
		},
		{
			input: "/razor-1911//",
			want:  "",
		},
		{
			input: "razor-1911-ampersand-skillion",
			want:  "Razor 1911 & Skillion",
		},
		{
			input: "razor-1911*trsi",
			want:  "Razor 1911, TRSi",
		},
		{
			input: "north-american-pirate_phreak-association",
			want:  "North American Pirate-Phreak Association",
		},
		{"2-minutes-to-midnight-bbs", "2 Minutes to Midnight BBS"},
		{"2000ad", "2000AD"},
		{"2tally-unrubbed", "2Tally Unrubbed"},
		{"2nd2none-bbs", "2ND2NONE BBS"},
		{"class*paradigm*razor-1911", "Class, Paradigm, Razor 1911"},
		{"down-town-bbs*bizare-bbs", "Down Town BBS, Bizare BBS"},
		{"united-software-association*fairlight", "United Software Association + Fairlight PC Division"},
		{"coop", "TDT / TRSi"},
	}

	for n, tt := range tests {
		t.Run("test humanize #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := releaser.Humanize(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "/home/ben/github/releaser",
			want:  "",
		},
		{
			input: "class",
			want:  "Class",
		},
		{
			input: "class*paradigm*razor-1911",
			want:  "Class + Paradigm + Razor 1911",
		},
		{
			input: "united-software-association*fairlight",
			want:  "United Software Association + Fairlight PC Division",
		},
		{
			input: "coop",
			want:  "TDT / TRSi",
		},
		{
			input: "razor-1911-demo*trsi",
			want:  "Razor 1911 Demo + TRSi",
		},
	}

	for n, tt := range tests {
		t.Run("test humanize #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := releaser.Link(tt.input)
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
		{"numbers only", "hello & world, foxes", "hello-ampersand-world*foxes"},
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
		{"special name", "coop", "TDT / TRSi"},
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
