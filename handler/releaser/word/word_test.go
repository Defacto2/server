package word_test

import (
	"testing"

	"github.com/Defacto2/server/handler/releaser/word"
	"github.com/nalgeon/be"
)

// When making changes to the package, make sure to run thorougher tests:
// go test . -cover -count=1000 -race

func TestAbbreviations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantAbbr string
		wantMix  string
	}{
		{"empty", "", "", ""},
		{"word", "modemS", "", ""},
		{"ordinal upper", "1ST", "1st", "1st"},
		{"ordinal lower", "1st", "1st", "1st"},
		{"ordinal mixed", "1St", "", "1st"},
		{"2nd mixed", "2nD", "", "2nd"},
		{"3rd mixed", "3Rd", "", "3rd"},
		{"iso lower", "iso", "ISO", "ISO"},
		{"iso upper", "ISO", "ISO", "ISO"},
		{"iso mixed", "Iso", "", "ISO"},
		{"gif lower", "gif", "GIF", "GIF"},
		{"uk lower", "uk", "UK", "UK"},
		{"whq mixed", "wHq", "", "WHQ"},
		{"bbs mixed", "bbS", "", "BBS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, word.Abbreviation(tt.input), tt.wantAbbr)
			be.Equal(t, word.AbbreviationMix(tt.input), tt.wantMix)
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
		{"no amp", "hello world", "hello world"},
		{"only amps", " & & ", ""},
		{"leading and trailing", "&hello&", "hello"},
		{"spaced", "hello & world", "hello & world"},
		{"gapless", "hello&world", "hello & world"},
		{"dupes", "hello&&world", "hello & world"},
		{"multiple", "a&b&c", "a & b & c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Amp(tt.input)
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
		{"no dots", "hello.", "HELLO"},
		{"comment 1", " Defacto2  demo  group. ", "DEFACTO2 DEMO GROUP"},
		{"comment 2", "the x bbs", "X BBS"},
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
		{"first", "the", 0, 5, ""},
		{"last", "the", 5, 5, ""},
		{"mixedcase", "ThE", 2, 5, ""},
		{"not a stop word", "foo", 2, 5, ""},
		{"lowercase", "of", 2, 5, "of"},
		{"titlecase", "The", 2, 5, "the"},
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

func TestEnglish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"lower", "hello", "Hello"},
		{"preserves inner caps", "mcDonald", "McDonald"},
		{"utf8", "café", "Café"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.English(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestFix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		position int
		last     int
		want     string
	}{
		{"connect", "of", 1, 3, "of"},
		{"abbreviation", "bbs", 1, 3, "BBS"},
		{"presuffix", "12am", 1, 3, "12AM"},
		{"sequence", "inc", 0, 3, "INC"},
		{"english", "world", 2, 2, "World"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Fix(tt.input, tt.position, tt.last)
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
		{"specific", "cybermail", "CyberMail"},
		{"dz", "hashx", "Hash X"},
		{"UPPER", "pcb", "PCB"},
		{"lower", "7Of9", "7of9"},
		{"exact upper", "Anz ftp", "ANZ FTP"},
		{"fmt by name", "Excretion anarchy", "eXCReTION"},
		{"am suffix", "the 12am group", "The 12AM Group"},
		{"pm suffix", "the 12pm group", "The 12PM Group"},
		{"dox", "thedox group", "TheDox Group"},
		{"fxp", "thefxp group", "TheFXP Group"},
		{"iso", "theiso group", "TheISO Group"},
		{"nfo", "thenfo group", "TheNFO Group"},
		{"pc", "pc-group", "PC-Group"},
		{"lsd", "the lsdgroup", "The LSDGroup"},
		{"inc", "inc group", "INC Group"},
		{"no dots", "hello.", "Hello"},
		{"example 1", "hello world.", "Hello World"},
		{"example 2", "the 12am group.", "The 12AM Group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Format(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestHyphen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no hyphen", "defacto2", ""},
		{"compound", "members-of-2000ad-will-meet-at-3pm", "Members-of-2000AD-Will-Meet-at-3PM"},
		{"boundary connect", "of-the", "Of-The"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Hyphen(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestPreSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short", "am", ""},
		{"am", "12Am", "12AM"},
		{"pm", "1pm", "1PM"},
		{"ad", "3000ad", "3000AD"},
		{"bc", "500bc", "500BC"},
		{"non-numeric am", "spam", ""},
		{"non-matching", "y2kadd", ""},
		{"dox", "razordox", "RazorDox"},
		{"fxp", "myfxp", "MyFXP"},
		{"iso", "defacto2iso", "Defacto2ISO"},
		{"nfo", "razor.nfo", "Razor.NFO"},
		{"pc prefix", "pc-now", "PC-Now"},
		{"lsd prefix", "lsdnow", "LSDNow"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.PreSuffix(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		index int
		want  string
	}{
		{"first inc", "inc", 0, "INC"},
		{"non-first inc", "inc", 1, ""},
		{"other word", "group", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.Sequence(tt.input, tt.index)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestStripChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"valid utf8", "ooÖØöøO", "ooÖØöøO"},
		{"symbols", "o.o|Ö+Ø=ö^ø#O", "ooÖØöøO"},
		{"accent and bang", "A Café!", "A Café"},
		{"hyphen utf8", "brunräven - över", "brunräven - över"},
		{"brackets", ".~[Hello]~.", "Hello"},
		{"dot", "defacto2.net", "defacto2net"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.StripChars(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestStripStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"clean", "hello world", "hello world"},
		{"hyphens", "--argument", "argument"},
		{"bangs", "!!!OMG-WTF", "OMG-WTF"},
		{"hash utf8", "#ÖØöøO", "ÖØöøO"},
		{"symbols", "!@#$%^&A(+)ooÖØöøO", "A(+)ooÖØöøO"},
		{"checkbox", " - [*] checkbox", "checkbox"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.StripStart(tt.input)
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
		{"utf8 accent with dot", "Café.", "Café"},
		{"utf8 accent no dot", "Café", "Café"},
		{"utf8 mixed", "Crème Brûlée.", "Crème Brûlée"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.TrimDot(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no spaces", "abc", "abc"},
		{"single spaces", "a b c", "a b c"},
		{"double spaces", "a  b  c", "a b c"},
		{"many spaces", "hello              world", "hello world"},
		{"control whitespace", "a\tb\nc\rd", "a b c d"},
		{"trailing tab", "a\t", "a "},
		{"single newline", "\n", " "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.TrimSP(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTrimThe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"short", "the", "the"},
		{"min length", "the bbs", "bbs"},
		{"bbs", "The X BBS", "X BBS"},
		{"ftp", "The X FTP", "X FTP"},
		{"lowercase", "the X BBS", "X BBS"},
		{"uppercase", "THE X BBS", "X BBS"},
		{"non-bbs", "The XXX", "The XXX"},
		{"no the", "Hello BBS", "Hello BBS"},
		{"multi word", "The High & Mighty Hello BBS", "High & Mighty Hello BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := word.TrimThe(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}
