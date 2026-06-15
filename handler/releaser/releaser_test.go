package releaser_test

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/initialism"
)

func listNames() []string {
	inits := *initialism.Initialisms()
	l := len(inits)
	n := make([]string, l)
	i := 0
	for k := range inits {
		n[i] = releaser.Humanize(string(k))
		i++
	}
	return n
}

func ExampleCell() {
	s := "  Defacto2  demo  group."
	fmt.Println(releaser.Cell(s))
	// Output: DEFACTO2 DEMO GROUP
}

func ExampleClean() {
	s := "  Defacto2  demo  group."
	fmt.Println(releaser.Clean(s))
	// Output: Defacto2 Demo Group
}

func ExampleHumanize() {
	path := "razor-1911-demo"
	fmt.Println(releaser.Humanize(path))
	// Output: Razor 1911 Demo
}

func ExampleLink() {
	path := "class*paradigm*razor-1911-demo"
	fmt.Println(releaser.Link(path))
	// Output: Class + Paradigm + Razor 1911 Demo
}

func ExampleObfuscate() {
	s := "Defacto2 Demo Group."
	fmt.Println(releaser.Obfuscate(s))
	// Output: defacto2-demo-group
}

func ExampleIndex() {
	fmt.Println(releaser.Index("united-software-association*fairlight"))
	fmt.Println(releaser.Index("class*paradigm*razor-1911"))
	fmt.Println(releaser.Index("coop"))
	// Output: UNITED SOFTWARE ASSOCIATION, FAIRLIGHT
	// CLASS, PARADIGM, RAZOR 1911
	// COOP
}

func BenchmarkCell(b *testing.B) {
	names := listNames()
	for b.Loop() {
		for n := range slices.Values(names) {
			if s := releaser.Cell(n); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkClean(b *testing.B) {
	names := listNames()
	for b.Loop() {
		for n := range slices.Values(names) {
			if s := releaser.Clean(n); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkHumanize(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Humanize(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkIndex(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Index(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkLink(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for uri := range *ins {
			path := releaser.Index(string(uri))
			if s := releaser.Link(path); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkObfuscate(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Obfuscate(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkTitle(b *testing.B) {
	for b.Loop() {
		for uri := range *initialism.Initialisms() {
			s := releaser.Index(string(uri))
			if title := releaser.Title(s); title != "" {
				_, _ = fmt.Fprintln(io.Discard, title)
			}
		}
	}
}

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
			if got := releaser.Cell(tt.args.s); got != strings.ToUpper(tt.want) {
				t.Errorf("Cell() = %v, want %v", got, strings.ToUpper(tt.want))
			}
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
			if got := releaser.Clean(tt.args.s); got != tt.want {
				t.Errorf("Clean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHumanize(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "defacto2",
			expected: "Defacto2",
		},
		{
			input:    "/razor-1911//",
			expected: "",
		},
		{
			input:    "razor-1911-ampersand-skillion",
			expected: "Razor 1911 & Skillion",
		},
		{
			input:    "razor-1911*trsi",
			expected: "Razor 1911, TRSi",
		},
		{
			input:    "north-american-pirate_phreak-association",
			expected: "North American Pirate-Phreak Association",
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

	for _, tc := range testCases {
		actual := releaser.Humanize(tc.input)
		if actual != tc.expected {
			t.Errorf("Humanize(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestLink(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "/home/ben/github/releaser",
			expected: "",
		},
		{
			input:    "class",
			expected: "Class",
		},
		{
			input:    "class*paradigm*razor-1911",
			expected: "Class + Paradigm + Razor 1911",
		},
		{
			input:    "united-software-association*fairlight",
			expected: "United Software Association + Fairlight PC Division",
		},
		{
			input:    "coop",
			expected: "TDT / TRSi",
		},
		{
			input:    "razor-1911-demo*trsi",
			expected: "Razor 1911 Demo + TRSi",
		},
	}

	for _, tc := range testCases {
		actual := releaser.Link(tc.input)
		if actual != tc.expected {
			t.Errorf("Link(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
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
			if got := releaser.Obfuscate(tt.arg); got != tt.want {
				t.Errorf("Obfuscate(%q) = %q, want %q", tt.arg, got, tt.want)
			}
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
			if got := releaser.Title(tt.arg); got != tt.want {
				t.Errorf("Title(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}
