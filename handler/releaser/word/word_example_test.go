package word_test

import (
	"fmt"
	"strings"

	"github.com/Defacto2/server/handler/releaser/word"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ExampleAbbreviation() {
	fmt.Println(word.Abbreviation("whq"))
	// Output: WHQ
}

func ExampleConnect() {
	titleize := cases.Title(language.English, cases.NoLower)
	const txt = "apple and oranges"
	s := strings.Split(titleize.String(txt), " ")
	for i, w := range s {
		x := word.Connect(w, i, len(s))
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
		x := word.Fix(w, i, len(s))
		if x != "" {
			s[i] = x
		}
	}
	fmt.Println(strings.Join(s, " "))
	// Output: Members of 2000AD Will Meet at 3PM
}

func ExampleHyphen() {
	s := "members-of-2000ad-will-meet-at-3pm"
	fmt.Println(word.Hyphen(s))
	// Output: Members-of-2000AD-Will-Meet-at-3PM
}

func ExampleFormat() {
	fmt.Println(word.Format("the BEST bbs"))
	// Output: The Best BBS
}

func ExampleStripChars() {
	fmt.Println(word.StripChars("!!!OMG-WTF???"))
	// Output: OMG-WTF
}

func ExampleTrimSP() {
	fmt.Print(word.TrimSP("            hello              world        "))
	// Output: hello world
}
