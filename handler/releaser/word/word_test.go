package word_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/releaser/word"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was at 95%+

// When making changes to the package, make sure to run thorougher tests:
// go test . -cover -count=1000 -race

func TestAbbreviations(t *testing.T) {
	t.Parallel()

	const s1 = "We the 1ST 2nD and 3Rd modemS to download the gif images from the uk wHq bbS !"
	s := strings.Split(s1, " ")
	for i, x := range s {
		got := word.Abbreviation(x)
		switch i {
		case 2:
			be.Equal(t, got, "1st")
		case 10:
			be.Equal(t, got, "GIF")
		case 14:
			be.Equal(t, got, "UK")
		default:
			be.Equal(t, got, "")
		}
	}

	for i, x := range s {
		got := word.AbbreviationMix(x)
		switch i {
		case 2:
			be.Equal(t, got, "1st")
		case 3:
			be.Equal(t, got, "2nd")
		case 5:
			be.Equal(t, got, "3rd")
		case 10:
			be.Equal(t, got, "GIF")
		case 14:
			be.Equal(t, got, "UK")
		case 15:
			be.Equal(t, got, "WHQ")
		case 16:
			be.Equal(t, got, "BBS")
		default:
			be.Equal(t, got, "")
		}
	}

	tests := [...]string{
		"1St", "1ST", "1st",
		"Iso", "iso", "ISO",
	}
	for n, s := range tests {
		mix := word.AbbreviationMix(s)
		switch n {
		case 0, 1, 2:
			be.Equal(t, mix, "1st")
		case 3, 4, 5:
			be.Equal(t, mix, "ISO")
		}

		got := word.Abbreviation(s)
		switch n {
		case 0:
			be.Equal(t, got, "")
		case 1, 2:
			be.Equal(t, got, "1st")
		case 3:
			be.Equal(t, got, "")
		case 4, 5:
			be.Equal(t, got, "ISO")
		}
	}
}

func TestTrimThe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"The X BBS", "X BBS"},
		{"The X FTP", "X FTP"},
		{"the X BBS", "X BBS"},
		{"THE X BBS", "X BBS"},
		{"The XXX", "The XXX"},
		{"Hello BBS", "Hello BBS"},
		{"The High & Mighty Hello BBS", "High & Mighty Hello BBS"},
	}
	for n, tt := range tests {
		t.Run("trim the #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := word.TrimThe(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimDot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no dots", "hello", "hello"},
		{"dot", "hello.", "hello"},
		{"dots", "hello..", "hello."},
		{"utf8_accent_with_dot", "Café.", "Café"},
		{"utf8_accent_no_dot", "Café", "Café"},
		{"utf8_mixed", "Crème Brûlée.", "Crème Brûlée"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.TrimDot(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestAmp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"str", "hello world", "hello world"},
		{"gap_amp", "hello & world", "hello & world"},
		{"gapless", "hello&world", "hello & world"},
		{"dupes", "hello&&world", "hello & world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Amp(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"EXACT", "beer", "BEER"},
		{"exact", "SceNET", "scenet"},
		{"specifc", "cybermail", "CyberMail"},
		{"dz", "hashx", "Hash X"},
		{"UPPER", "pcb", "PCB"},
		{"lower", "7Of9", "7of9"},
		{"exact_upper", "Anz ftp", "ANZ FTP"},
		{"fmt_by_name", "Excretion anarchy", "eXCReTION"},
		{"am_suffix", "the 12am group", "The 12AM Group"},
		{"pm_suffix", "the 12pm group", "The 12PM Group"},
		{"dox", "thedox group", "TheDox Group"},
		{"fxp", "thefxp group", "TheFXP Group"},
		{"iso", "theiso group", "TheISO Group"},
		{"nfo", "thenfo group", "TheNFO Group"},
		{"pc", "pc-group", "PC-Group"},
		{"lsd", "the lsdgroup", "The LSDGroup"},
		{"inc", "inc group", "INC Group"},
		{"no_dots", "hello.", "Hello"},
		{"example_1", "hello world.", "Hello World"},
		{"example_2", "the 12am group.", "The 12AM Group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Format(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestCell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"EXACT", "beer", "BEER"},
		{"exact", "SceNET", "SCENET"},
		{"pc", "pc-group", "PC-GROUP"},
		{"no_dots", "hello.", "HELLO"},
		{"comment_1", " Defacto2  demo  group. ", "DEFACTO2 DEMO GROUP"},
		{"comment_2", "the x bbs", "X BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Cell(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		word     string
		position int
		last     int
		want     string
	}{
		{"empty", "", 0, 0, ""},
		{"first", "hello", 0, 5, ""},
		{"last", "world", 4, 5, ""},
		{"mixedcase", "ThE", 2, 5, ""},
		{"not a stop word", "foo", 2, 5, ""},
		{"lowercase", "of", 2, 5, "of"},
		{"uppercase", "THE", 2, 5, "the"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Connect(tt.word, tt.position, tt.last)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestStripChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"ooÖØöøO", "ooÖØöøO"},
		{"o.o|Ö+Ø=ö^ø#O", "ooÖØöøO"},
		{"A Café!", "A Café"},
		{"brunräven - över", "brunräven - över"},
		{".~[Hello]~.", "Hello"},
		{"defacto2.net", "defacto2net"},
	}
	for n, tt := range tests {
		t.Run("string chars #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := word.StripChars(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestStripStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"hello world", "hello world"},
		{"--argument", "argument"},
		{"!!!OMG-WTF", "OMG-WTF"},
		{"#ÖØöøO", "ÖØöøO"},
		{"!@#$%^&A(+)ooÖØöøO", "A(+)ooÖØöøO"},
		{" - [*] checkbox", "checkbox"},
	}
	for n, tt := range tests {
		t.Run("strip start #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := word.StripStart(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"abc", "abc"},
		{"a b c", "a b c"},
		{"a  b  c", "a b c"},
		{"hello              world", "hello world"},
	}
	for n, tt := range tests {
		t.Run("trimsp #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := word.TrimSP(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestPreSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"12Am", "12AM"},
		{"1pm", "1PM"},
		{"3000ad", "3000AD"},
		{"y2kadd", ""},
		{"razordox", "RazorDox"},
		{"myfxp", "MyFXP"},
		{"defacto2iso", "Defacto2ISO"},
		{"razor.nfo", "Razor.NFO"},
		{"pc-now", "PC-Now"},
		{"lsdnow", "LSDNow"},
	}
	for n, tt := range tests {
		t.Run("presuffix #"+strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			got := word.PreSuffix(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}
