package fix_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/releaser/fix"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ExampleAbbreviation() {
	fmt.Println(fix.Abbreviation("whq"))
	// Output: WHQ
}

func ExampleConnect() {
	titleize := cases.Title(language.English, cases.NoLower)
	const txt = "apple and oranges"
	s := strings.Split(titleize.String(txt), " ")
	for i, w := range s {
		x := fix.Connect(w, i, len(s))
		if x != "" {
			s[i] = x
		}
	}
	fmt.Println(strings.Join(s, " "))
	// Output: Apple and Oranges
}

func ExampleFix() {
	titleize := cases.Title(language.English, cases.NoLower)
	const txt = "members of 2000ad will meet at 3pm"
	s := strings.Split(titleize.String(txt), " ")
	for i, w := range s {
		x := fix.Fix(w, i, len(s))
		if x != "" {
			s[i] = x
		}
	}
	fmt.Println(strings.Join(s, " "))
	// Output: Members of 2000AD Will Meet at 3PM
}

func ExampleHyphen() {
	s := "members-of-2000ad-will-meet-at-3pm"
	fmt.Println(fix.Hyphen(s))
	// Output: Members-of-2000AD-Will-Meet-at-3PM
}

func ExampleFormat() {
	fmt.Println(fix.Format("the BEST bbs"))
	// Output: The Best BBS
}

func ExampleStripChars() {
	fmt.Println(fix.StripChars("!!!OMG-WTF???"))
	// Output: OMG-WTF
}

func ExampleTrimSP() {
	fmt.Print(fix.TrimSP("            hello              world        "))
	// Output: hello world
}

func TestTrimThe(t *testing.T) {
	t.Parallel()
	type args struct {
		g string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"The X BBS"}, "X BBS"},
		{"", args{"The X FTP"}, "X FTP"},
		{"", args{"the X BBS"}, "X BBS"},
		{"", args{"THE X BBS"}, "X BBS"},
		{"", args{"The"}, "The"},
		{"", args{"Hello BBS"}, "Hello BBS"},
		{"", args{"The High & Mighty Hello BBS"}, "High & Mighty Hello BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.TrimThe(tt.args.g); got != tt.want {
				t.Errorf("TrimThe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimDot(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"no dots", args{"hello"}, "hello"},
		{"dot", args{"hello."}, "hello"},
		{"dots", args{"hello.."}, "hello."},
		{"utf8_accent_with_dot", args{"Café."}, "Café"},
		{"utf8_accent_no_dot", args{"Café"}, "Café"},
		{"utf8_mixed", args{"Crème Brûlée."}, "Crème Brûlée"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.TrimDot(tt.args.s); got != tt.want {
				t.Errorf("TrimDot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAmp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		w    string
		want string
	}{
		{"empty", "", ""},
		{"str", "hello world", "hello world"},
		{"gap amp", "hello & world", "hello & world"},
		{"gapless", "hello&world", "hello & world"},
		{"dupes", "hello&&world", "hello & world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.Amp(tt.w); got != tt.want {
				t.Errorf("Amp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"EXACT", "beer", "BEER"},
		{"exact", "SceNET", "scenet"},
		{"specifc", "cybermail", "CyberMail"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.Format(tt.s); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"EXACT", "beer", "BEER"},
		{"exact", "SceNET", "SCENET"},
		{"pc", "pc-group", "PC-GROUP"},
		{"no dots", "hello.", "HELLO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.Cell(tt.s); got != tt.want {
				t.Errorf("Cell() = %v, want %v", got, tt.want)
			}
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
		{"lowercase", "of", 2, 5, "of"},
		{"uppercase", "THE", 2, 5, "the"},
		{"mixedcase", "ThE", 2, 5, "the"},
		{"not a stop word", "foo", 2, 5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.Connect(tt.word, tt.position, tt.last); got != tt.want {
				t.Errorf("Connect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StripChars(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"ooÖØöøO"}, "ooÖØöøO"},
		{"", args{"o.o|Ö+Ø=ö^ø#O"}, "ooÖØöøO"},
		{"", args{"A Café!"}, "A Café"},
		{"", args{"brunräven - över"}, "brunräven - över"},
		{"", args{".~[Hello]~."}, "Hello"},
		{"", args{"defacto2.net"}, "defacto2net"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.StripChars(tt.args.s); got != tt.want {
				t.Errorf("StripChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StripStart(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"hello world"}, "hello world"},
		{"", args{"--argument"}, "argument"},
		{"", args{"!!!OMG-WTF"}, "OMG-WTF"},
		{"", args{"#ÖØöøO"}, "ÖØöøO"},
		{"", args{"!@#$%^&A(+)ooÖØöøO"}, "A(+)ooÖØöøO"},
		{"", args{" - [*] checkbox"}, "checkbox"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.StripStart(tt.args.s); got != tt.want {
				t.Errorf("StripStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TrimSP(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"abc"}, "abc"},
		{"", args{"a b c"}, "a b c"},
		{"", args{"a  b  c"}, "a b c"},
		{"", args{"hello              world"}, "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fix.TrimSP(tt.args.s); got != tt.want {
				t.Errorf("TrimSP() = %v, want %v", got, tt.want)
			}
		})
	}
}
