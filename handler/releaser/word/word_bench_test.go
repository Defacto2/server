package word_test

import (
	"testing"

	"github.com/Defacto2/server/handler/releaser/word"
)

func BenchmarkAbbreviation(b *testing.B) {
	inputs := []string{"ISO", "iso", "1st", "example", "running", "BBS", "something"}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Abbreviation(s)
		}
	}
}

func BenchmarkAbbreviationMix(b *testing.B) {
	inputs := []string{"ISO", "iso", "Iso", "1St", "example", "wHq", "bbS"}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.AbbreviationMix(s)
		}
	}
}

func BenchmarkAmp(b *testing.B) {
	inputs := []string{
		"hello world",
		"  &&hello&&world&&foo&bar&&  ",
		"hello & world",
		"a&b&c",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Amp(s)
		}
	}
}

func BenchmarkCell(b *testing.B) {
	inputs := []string{
		" Defacto2  demo  group. , the x bbs",
		"Lightforce,Pact,TRSi,Venom,Razor 1911,the System",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Cell(s)
		}
	}
}

func BenchmarkConnect(b *testing.B) {
	inputs := []string{"the", "THE", "The", "of", "and", "defacto2"}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Connect(s, 1, 5)
		}
	}
}

func BenchmarkEnglish(b *testing.B) {
	inputs := []string{"hello", "mcDonald", "café", "defacto2"}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.English(s)
		}
	}
}

func BenchmarkFix(b *testing.B) {
	inputs := []string{"the", "bbs", "12am", "inc", "world", "razordox"}
	last := len(inputs) - 1
	for b.Loop() {
		for i, s := range inputs {
			_ = word.Fix(s, i, last)
		}
	}
}

func BenchmarkFormat(b *testing.B) {
	inputs := []string{
		"the 12am group.",
		"Excretion anarchy",
		"Lightforce,Pact,TRSi,Venom,Razor 1911,the System",
		"pc-group",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Format(s)
		}
	}
}

func BenchmarkHyphen(b *testing.B) {
	inputs := []string{
		"nohyphen",
		"multi-part-hyphenated-word-test",
		"members-of-2000ad-will-meet-at-3pm",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.Hyphen(s)
		}
	}
}

func BenchmarkPreSuffix(b *testing.B) {
	inputs := []string{"12am", "superdox", "pc-game", "lsdnow", "regularword"}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.PreSuffix(s)
		}
	}
}

func BenchmarkSequence(b *testing.B) {
	inputs := []string{"inc", "group", "bbs"}
	for b.Loop() {
		for i, s := range inputs {
			_ = word.Sequence(s, i)
		}
	}
}

func BenchmarkStripChars(b *testing.B) {
	inputs := []string{
		"Café!",
		".~[[@]hello[@]]~.",
		"Clean Releaser Name",
		"123-ABC, & Co.",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.StripChars(s)
		}
	}
}

func BenchmarkStripStart(b *testing.B) {
	inputs := []string{
		" - [*] checkbox",
		"checkbox",
		"*** 123 test",
		"--- !!! ---",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.StripStart(s)
		}
	}
}

func BenchmarkTrimDot(b *testing.B) {
	inputs := []string{"hello", "hello.", "hello..", "Café."}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.TrimDot(s)
		}
	}
}

func BenchmarkTrimSP(b *testing.B) {
	inputs := []string{
		"hello              world",
		"hello world",
		"  multiple   spaces   between   words  ",
		"a\tb\nc\r",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.TrimSP(s)
		}
	}
}

func BenchmarkTrimThe(b *testing.B) {
	inputs := []string{
		"The X BBS",
		"the super cool ftp",
		"The X",
		"X BBS",
		"Short",
	}
	for b.Loop() {
		for _, s := range inputs {
			_ = word.TrimThe(s)
		}
	}
}
