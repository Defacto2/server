// Package word provides functions for cleaning and formatting strings of known words and group names.
package word

import (
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/Defacto2/server/handler/releaser/name"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const space = " "

func init() {
	// initializing maps of key values on startup is the most performant
	// method of handing search and replace keywords. Importantly, using a
	// preallocated map nullifies RAM consumption when running a query.

	rawAbbr := map[string]string{
		// to lowercase
		"1st": "1st", "2nd": "2nd", "3rd": "3rd", "4th": "4th", "5th": "5th",
		"6th": "6th", "7th": "7th", "8th": "8th", "9th": "9th", "10th": "10th",
		"11th": "11th", "12th": "12th", "13th": "13th", "7of9": "7of9",
		// to uppercase
		"3d": "3D", "abc": "ABC", "acdc": "ACDC", "ad": "AD", "am": "AM",
		"amf": "AMF", "ansi": "ANSI", "asm": "ASM", "au": "AU", "bbc": "BBC",
		"bbs": "BBS", "bc": "BC", "cd": "CD", "cgi": "CGI", "diz": "DIZ",
		"dox": "DOX", "eu": "EU", "faq": "FAQ", "fbi": "FBI", "fm": "FM",
		"ftp": "FTP", "fr": "FR", "fx": "FX", "fxp": "FXP", "gbc": "GBC",
		"gif": "GIF", "hq": "HQ", "id": "ID", "ii": "II", "iii": "III",
		"iso": "ISO", "kgb": "KGB", "mp3": "MP3", "pc": "PC", "pcb": "PCB",
		"pcp": "PCP", "pda": "PDA", "pm": "PM", "psx": "PSX", "pwa": "PWA",
		"rom": "ROM", "rpm": "RPM", "ssd": "SSD", "st": "ST", "tnt": "TNT",
		"tsr": "TSR", "ufo": "UFO", "uk": "UK", "us": "US", "usa": "USA",
		"uss": "USS", "ussr": "USSR", "vcd": "VCD", "whq": "WHQ", "xxx": "XXX",
	}

	rawConn := [...]string{
		"a", "as", "and", "at", "by", "el", "of", "for", "from", "in", "is", "or", "tha",
		"the", "to", "with",
	}

	connectingWords = make(map[string]string, len(rawConn)*3)
	for _, s := range rawConn {
		connectingWords[s] = s
		connectingWords[strings.ToUpper(s)] = s
		connectingWords[strings.ToUpper(s[:1])+s[1:]] = s
	}

	abbreviations = make(map[string]string, len(rawAbbr)*2)
	for key, v := range rawAbbr {
		abbreviations[key] = v
		abbreviations[strings.ToUpper(key)] = v
	}
}

var (
	abbreviations   map[string]string
	connectingWords map[string]string
)

var engTitlePool = sync.Pool{
	// a sync.pool is required due to an occasional and random panic,
	// caused by out-of-range indexing
	New: func() any {
		c := cases.Title(language.English, cases.NoLower)
		return &c
	},
}

// English is intended to safely return an English title using upper casing lead.
func English(s string) string {
	caserPtr := engTitlePool.Get().(*cases.Caser)
	defer engTitlePool.Put(caserPtr)

	return (*caserPtr).String(s)
}

// Abbreviation applies upper casing to known acronyms, initialisms and abbreviations.
// It applies lower casing to ordinal numbers 1st through to 13th.
// Otherwise it returns an empty string.
//
// It requires lowercase input, otherwise use the slower [AbbreviationMix].
//
// Example:
//
//	Abbreviation("1ST") = "1st"
//	Abbreviation("1sT") = ""
//	Abbreviation("iso") = "ISO"
//	Abbreviation("Iso") = ""
func Abbreviation(lcase string) string {
	return abbreviations[lcase]
}

// Abbreviation applies upper casing to known acronyms, initialisms and abbreviations.
// It applies lower casing to ordinal numbers 1st through to 13th.
// Otherwise it returns an empty string.
//
// It supports mixed case input,
// however if s is confirmed to be lowercase, use the 2x faster [Abbreviation].
//
// Example:
//
//	Abbreviation("1ST") = "1st"
//	Abbreviation("1sT") = "1st"
//	Abbreviation("iso") = "ISO"
//	Abbreviation("Iso") = "ISO"
func AbbreviationMix(s string) string {
	if res, ok := abbreviations[s]; ok {
		return res
	}
	// fallback for mix casing: "Iso", "1St"
	return abbreviations[strings.ToLower(s)]
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

	s = strings.Trim(s, "& ")
	if s == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(s) + 4)

	var inAmp bool
	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '&' {
			if !inAmp {
				inAmp = true
				if builder.Len() > 0 && builder.String()[builder.Len()-1] != ' ' {
					builder.WriteByte(' ')
				}
				builder.WriteByte('&')
			}
			continue
		}

		if inAmp {
			inAmp = false
			if ch != ' ' {
				builder.WriteByte(' ')
			}
		}
		builder.WriteByte(ch)
	}

	return builder.String()
}

// Connect formats common connecting words based on the position in a slice.
// The goal is to allow for US style titles: The Best BBS and the Rest.
//
// For optimal performance, it does not support mixed case strings.
//
// Example:
//
//	Connect("at", 1, 99) = "at"
//	Connect("AT", 1, 99) = "at"
//	Connect("at", 0, 99) = ""
//	Connect("at", 99, 99) = ""
//	Connect("common", 8, 99) = ""
func Connect(s string, position, last int) string {
	if position == 0 || position == last {
		return ""
	}
	return connectingWords[s]
}

// Cell returns a copy of s with custom formatting for storage in a database cell.
// All words will be upper cased and stripped of incompatible characters.
//
// Example:
//
//	Cell(" Defacto2  demo  group. ") = "DEFACTO2 DEMO GROUP"
//	Cell("the x bbs") = "X BBS"
func Cell(s string) string {
	groups := strings.Split(s, ",")

	for i, group := range groups {
		group = TrimThe(TrimSP(group))
		lcase := strings.ToLower(strings.TrimSpace(group))
		lcase = Amp(lcase)
		elems := strings.Split(lcase, space)
		last := len(elems) - 1

		for n, elem := range elems {
			elem = TrimDot(elem)
			if fix := Hyphen(elem); fix != "" {
				elems[n] = fix
				continue
			}
			elems[n] = Fix(elem, n, last)
		}
		groups[i] = strings.Join(elems, space)
	}

	return strings.ToUpper(strings.Join(groups, ", "))
}

// Fix formats the s string based on its position in the words slice.
// The position is the index of the word in the words slice.
// The last is the index of the last word in the words slice.
func Fix(s string, position, last int) string {
	if fix := Connect(s, position, last); fix != "" {
		return fix
	}

	if fix := Abbreviation(s); fix != "" {
		return fix
	}

	if fix := PreSuffix(s); fix != "" {
		return fix
	}

	if fix := Sequence(s, position); fix != "" {
		return fix
	}

	return English(s)
}

// Hyphen replaces spaces in a sentence with hyphens.
// Each word is also parsed using [fix.Fix].
//
// Example:
//
//	Hyphen("members of 2000ad will meet at 3pm") = "Members-of-2000AD-Will-Meet-at-3PM"
func Hyphen(s string) string {
	const hyphen = "-"
	if !strings.Contains(s, hyphen) {
		return ""
	}

	compounds := strings.Split(s, hyphen)
	last := len(compounds) - 1
	for i, compound := range compounds {
		compounds[i] = Fix(compound, i, last)
	}

	return strings.Join(compounds, hyphen)
}

// Format returns a copy of s with custom formatting.
// Some words and known acronyms will be upper cased, lower cased or title cased.
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

		elems := strings.Split(fullname, space)
		last := len(elems) - 1
		for i, elem := range elems {
			elem = TrimDot(elem)
			if fix := Hyphen(elem); fix != "" {
				elems[i] = fix
				continue
			}
			elems[i] = Fix(elem, i, last)
		}

		groups[index] = strings.Join(elems, space)
	}

	return strings.Join(groups, ", ")
}

// PreSuffix formats the string if a known prefix or suffix is found.
// The title caser needs to be a language-specific title casing.
//
// Example:
//
//	PreSuffix("12am") = "12AM"
func PreSuffix(s string) string {
	if len(s) < 3 {
		return ""
	}
	suff := s[len(s)-2:]
	if suff == "am" || suff == "AM" || suff == "Am" ||
		suff == "pm" || suff == "PM" || suff == "Pm" ||
		suff == "ad" || suff == "AD" || suff == "Ad" ||
		suff == "bc" || suff == "BC" || suff == "Bc" {

		prefix := s[:len(s)-2]
		if _, err := strconv.Atoi(prefix); err == nil {
			return prefix + strings.ToUpper(suff)
		}
	}

	s = strings.ToLower(s)
	switch {
	case strings.HasSuffix(s, "dox"):
		return English(strings.TrimSuffix(s, "dox")) + "Dox"
	case strings.HasSuffix(s, "fxp"):
		return English(strings.TrimSuffix(s, "fxp")) + "FXP"
	case strings.HasSuffix(s, "iso"):
		return English(strings.TrimSuffix(s, "iso")) + "ISO"
	case strings.HasSuffix(s, "nfo"):
		return English(strings.TrimSuffix(s, "nfo")) + "NFO"
	case strings.HasPrefix(s, "pc-"):
		return "PC-" + English(strings.TrimPrefix(s, "pc-"))
	case strings.HasPrefix(s, "lsd"):
		return "LSD" + English(strings.TrimPrefix(s, "lsd"))
	}
	return ""
}

// Sequence formats the string when i is 0, meaning s is the first word.
func Sequence(s string, i int) string {
	if i != 0 {
		return ""
	}
	switch s {
	case "inc":
		return "INC"
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
	// this manual implementation is around 6x faster than using a regular expression,
	// and uses around 14x less memory.

	remove := false
	for _, r := range s {
		if !valid(r) {
			remove = true
			break
		}
	}
	if !remove {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if valid(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func valid(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	switch r {
	case '-', ',', '&', ' ':
		return true
	}
	return false
}

// StripStart removes the non-alphanumeric characters from the start of the string.
//
// Example:
//
//	StripStart(" - [*] checkbox") = "checkbox"
func StripStart(s string) string {
	// this is 10x more performant than a regex and uses 0 memory allocation
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return s[i:]
		}
	}
	return ""
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
	// using this manual replacement pattern has been benchmarked
	// with offering 10x the performance of a regular expression.
	if s == "" {
		return ""
	}
	hasDupes := false
	for i := 0; i < len(s)-1; i++ {
		if (s[i] == ' ' && s[i+1] == ' ') || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' {
			hasDupes = true
			break
		}
	}
	if !hasDupes {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s))
	inSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if !inSpace {
				builder.WriteByte(' ')
				inSpace = true
			}
		} else {
			builder.WriteByte(c)
			inSpace = false
		}
	}
	return builder.String()
}

// TrimThe drops "The " prefix whenever the named string ends with " BBS" or " FTP".
// It is to avoid unique site names duplication, e.g. "The X BBS" and "X BBS".
//
// Example:
//
//	TrimThe("The X BBS") = "X BBS"
//	TrimThe("X BBS") = "X BBS"
//	TrimThe("The XXX") = "The XXX" // no change
func TrimThe(s string) string {
	if len(s) < 7 { // "the bbs" || "the ftp" are minimum requirements
		return s
	}

	if !strings.EqualFold(s[:4], "the ") {
		return s
	}

	if strings.HasSuffix(strings.ToUpper(s), " BBS") ||
		strings.HasSuffix(strings.ToUpper(s), " FTP") {
		return strings.TrimSpace(s[4:])
	}

	return s
}
