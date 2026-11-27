package word_test

import (
	"fmt"
	"strings"

	"github.com/Defacto2/server/handler/releaser/word"
)

func ExampleAbbreviation() {
	fmt.Println(word.Abbreviation("whq"))
	// Output: WHQ
}

func ExampleConnect() {
	s := strings.Split(word.English("apple and oranges"), " ")
	last := len(s) - 1
	for i, w := range s {
		if x := word.Connect(w, i, last); x != "" {
			s[i] = x
		}
	}
	fmt.Println(strings.Join(s, " "))
	// Output: Apple and Oranges
}

func ExampleFix() {
	s := strings.Split("members of 2000ad will meet at 3pm", " ")
	last := len(s) - 1
	for i, w := range s {
		s[i] = word.Fix(w, i, last)
	}
	fmt.Println(strings.Join(s, " "))
	// Output: Members of 2000AD Will Meet at 3PM
}

func ExampleFormat() {
	fmt.Println(word.Format("the BEST bbs"))
	// Output: The Best BBS
}

func ExampleHyphen() {
	s := "members-of-2000ad-will-meet-at-3pm"
	fmt.Println(word.Hyphen(s))
	// Output: Members-of-2000AD-Will-Meet-at-3PM
}

func ExampleStripChars() {
	fmt.Println(word.StripChars("!!!OMG-WTF???"))
	// Output: OMG-WTF
}

func ExampleTrimSP() {
	fmt.Println(word.TrimSP("hello              world"))
	// Output: hello world
}
