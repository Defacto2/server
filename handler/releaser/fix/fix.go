// Package fix provides functions for cleaning and formatting strings of known words and group names.
package fix

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/releaser/name"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const space = " "

// Abbreviation applies upper casing to known acronyms, initialisms and abbreviations.
// And lower casing to ordinal numbers 1st through to 13th.
// Otherwise it returns an empty string.
//
// Example:
//
//	Abbreviation("1ST") = "1st"
//	Abbreviation("iso") = "ISO"
func Abbreviation(s string) string {
	x := strings.ToLower(s)
	switch x {
	case "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th",
		"10th", "11th", "12th", "13th":
		return strings.ToLower(s)
	case "3d", "abc", "acdc", "ad", "am", "amf", "ansi", "asm", "au", "bbc", "bbs", "bc",
		"cd", "cgi", "diz", "dox", "eu", "faq", "fbi", "fm", "ftp", "fr", "fx", "fxp",
		"gbc", "gif", "hq", "id", "ii", "iii", "iso", "kgb", "mp3", "pc", "pcb", "pcp",
		"pda", "pm", "psx", "pwa", "rom", "rpm", "ssd", "st", "tnt", "tsr", "ufo", "uk",
		"us", "usa", "uss", "ussr", "vcd", "whq", "xxx":
		return strings.ToUpper(s)
	case "7of9":
		return strings.ToLower(s)
	default:
		return ""
	}
}

// Amp formats the special ampersand (&) character in the string
// to be usable with a URL path in use by the group.
//
// Example:
//
//	Amp("hello&&world") = "hello & world"
func Amp(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	x := s
	trimDupes := regexp.MustCompile(`\&+`)
	x = trimDupes.ReplaceAllString(x, "&")

	trimPrefix := regexp.MustCompile(`^\&+`)
	x = trimPrefix.ReplaceAllString(x, "")

	trimSuffix := regexp.MustCompile(`\&+$`)
	x = trimSuffix.ReplaceAllString(x, "")

	addWhitespace := regexp.MustCompile(`(\S)\&(\S)`) // \S matches any character that's not whitespace
	x = addWhitespace.ReplaceAllString(x, "$1 & $2")
	return x
}

// Connect formats common connecting word as the w string based on its position in a words slice.
func Connect(w string, position, last int) string {
	const first = 0
	if position == first || position == last {
		return ""
	}
	switch strings.ToLower(w) {
	case "a", "as", "and", "at", "by", "el", "of", "for", "from", "in", "is", "or", "tha",
		"the", "to", "with":
		return strings.ToLower(w)
	}
	return ""
}

// Cell returns a copy of s with custom formatting for storage in a database cell.
// All words will be upper cased and stipped of incompatible characters.
//
// Example:
//
//	Cell(" Defacto2  demo  group. ") = "DEFACTO2 DEMO GROUP"
//	Cell("the x bbs") = "X BBS"
func Cell(s string) string {
	groups := strings.Split(s, ",")
	for index, group := range groups {
		fullname := strings.ToLower(strings.TrimSpace(group))
		fullname = Amp(fullname)
		words := strings.Split(fullname, space)
		last := len(words) - 1
		for i, word := range words {
			word = TrimDot(word)
			if fix := Hyphen(word); fix != "" {
				words[i] = fix
				continue
			}
			words[i] = Fix(word, i, last)
		}
		groups[index] = strings.Join(words, space)
	}
	return strings.ToUpper(strings.Join(groups, ", "))
}

// Fix formats the w string based on its position in the words slice.
// The position is the index of the word in the words slice.
// The last is the index of the last word in the words slice.
func Fix(w string, position, last int) string {
	if fix := Connect(w, position, last); fix != "" {
		return fix
	}
	if fix := Abbreviation(w); fix != "" {
		return fix
	}
	title := cases.Title(language.English, cases.NoLower)
	if fix := PreSuffix(w, title); fix != "" {
		return fix
	}
	if fix := Sequence(w, position); fix != "" {
		return fix
	}
	return title.String(w)
}

// Hyphen applies [fix.Fix] to hyphenated words.
func Hyphen(w string) string {
	const hyphen = "-"
	if !strings.Contains(w, hyphen) {
		return ""
	}
	compounds := strings.Split(w, hyphen)
	last := len(compounds) - 1
	for i, word := range compounds {
		compounds[i] = Fix(word, i, last)
	}
	return strings.Join(compounds, hyphen)
}

// Format returns a copy of s with custom formatting.
// Certain words and known acronyms will be upper cased, lower cased or title cased.
// Known named groups will be returned in their special casing.
// Trailing dots will be removed.
//
// Example:
//
//	Format("hello world.") = "Hello World"
//	Format("the 12am group.") = "The 12AM Group"
func Format(s string) string {
	const acronym = 3
	if len(s) <= acronym {
		return strings.ToUpper(s)
	}
	groups := strings.Split(s, ",")
	for index, group := range groups {
		fullname := strings.ToLower(strings.TrimSpace(group))
		fullname = Amp(fullname)
		if special := name.Obfuscate(fullname).String(); special != "" {
			groups[index] = special
			continue
		}
		words := strings.Split(fullname, space)
		last := len(words) - 1
		for i, word := range words {
			word = TrimDot(word)
			if fix := Hyphen(word); fix != "" {
				words[i] = fix
				continue
			}
			words[i] = Fix(word, i, last)
		}
		groups[index] = strings.Join(words, space)
	}
	return strings.Join(groups, ", ")
}

// PreSuffix formats the w string if a known prefix or suffix is found.
// The title caser needs to be a language-specific title casing.
//
// Example:
//
//	PreSuffix("12am", cases.Title(language.English, cases.NoLower)) = "12AM"
func PreSuffix(s string, title cases.Caser) string {
	word := strings.ToLower(s)
	atois := []string{"ad", "bc", "am", "pm"}
	for _, suffix := range atois {
		if !strings.HasSuffix(word, suffix) {
			continue
		}
		trim := strings.TrimSuffix(word, suffix)
		value, err := strconv.Atoi(trim)
		if err != nil {
			continue
		}
		return fmt.Sprintf("%d%s", value, strings.ToUpper(suffix))
	}
	switch {
	case strings.HasSuffix(word, "dox"):
		return title.String(strings.TrimSuffix(word, "dox")) + "Dox"
	case strings.HasSuffix(word, "fxp"):
		return title.String(strings.TrimSuffix(word, "fxp")) + "FXP"
	case strings.HasSuffix(word, "iso"):
		return title.String(strings.TrimSuffix(word, "iso")) + "ISO"
	case strings.HasSuffix(word, "nfo"):
		return title.String(strings.TrimSuffix(word, "nfo")) + "NFO"
	case strings.HasPrefix(word, "pc-"):
		return "PC-" + title.String(strings.TrimPrefix(word, "pc-"))
	case strings.HasPrefix(word, "lsd"):
		return "LSD" + title.String(strings.TrimPrefix(word, "lsd"))
	}
	return ""
}

// Sequence formats the w string if it is the first word in the words slice.
func Sequence(w string, i int) string {
	if i != 0 {
		return ""
	}
	switch w { //nolint:gocritic
	case "inc":
		// note: Format() applies UPPER to all 3 letter or smaller words
		return strings.ToUpper(w)
	}
	return ""
}

// StripChars removes all the incompatible characters that cannot be used for releaser URL paths.
//
// Example:
//
//	StripChars("Café!") = "Café"
//	StripChars(".~[[@]hello[@]]~.") = "hello"
func StripChars(s string) string {
	const validChars = `[^A-Za-zÀ-ÖØ-öø-ÿ0-9\-,& ]`
	r := regexp.MustCompile(validChars)
	return r.ReplaceAllString(s, "")
}

// StripStart removes the non-alphanumeric characters from the start of the string.
//
// Example:
//
//	StripStart(" - [*] checkbox") = "checkbox"
func StripStart(s string) string {
	const latinChars = `[A-Za-z0-9À-ÖØ-öø-ÿ]`
	r := regexp.MustCompile(latinChars)
	f := r.FindStringIndex(s)
	if f == nil {
		return ""
	}
	if f[0] != 0 {
		return s[f[0]:]
	}
	return s
}

// TrimDot removes a trailing dot from s.
//
// Example:
//
//	TrimDot("hello.") = "hello"
//	TrimDot("hello..") = "hello."
func TrimDot(s string) string {
	t, ok := strings.CutSuffix(s, ".")
	if ok {
		return t
	}
	return s
}

// TrimSP removes duplicate spaces from the string.
//
// Example:
//
//	TrimSP("hello              world") = "hello world"
func TrimSP(s string) string {
	const spaces = `\s+`
	r := regexp.MustCompile(spaces)
	return r.ReplaceAllString(s, " ")
}

// TrimThe drops "The " prefix whenever the named string ends with " BBS" or " FTP".
// It is to avoid unique site names duplication, e.g. "The X BBS" and "X BBS".
//
// Example:
//
//	TrimThe("The X BBS") = "X BBS"
//	TrimThe("X BBS") = "X BBS"
//	TrimThe("The X") = "The X" // no change
func TrimThe(name string) string {
	const short = 2
	a := strings.Split(name, space)
	if len(a) < short {
		return name
	}
	l := strings.ToUpper(a[len(a)-1])
	if strings.EqualFold(a[0], "the") && (l == "BBS" || l == "FTP") {
		return strings.TrimSpace(strings.Join(a[1:], space)) // drop "the" prefix
	}
	return name
}
