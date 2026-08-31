package word_test

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const space = " "

// These benchmarks were used for potential methods of dealing with single word,
// and sentence, search and replace through a document.
// Benchmarks ending with 00 were the original unoptimized implementations
// in use until September 2026.
//
// Use the following command to run all:
// go test -bench=Benchmark -benchmem

// ABBREVIATIONS
//
// go test -bench=BenchmarkAbbr -benchmem

func BenchmarkAbbr00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr00(w)
		}
	}
}

func BenchmarkAbbr01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr01(w)
		}
	}
}

func BenchmarkAbbr02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr02(w)
		}
	}
}

func BenchmarkAbbr03(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr03(w)
		}
	}
}

func BenchmarkAbbr04(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr04(w)
		}
	}
}

func BenchmarkAbbr05(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr05(w)
		}
	}
}

func BenchmarkAbbr06(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, w := range testAbbr {
			_ = abbr06(w)
		}
	}
}

// AMP
//
// go test -bench=BenchmarkAmp -benchmem

func BenchmarkAmp00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = amp00(testAmp)
	}
}

func BenchmarkAmp01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = amp01(testAmp)
	}
}

func BenchmarkAmp02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = amp02(testAmp)
	}
}

// CELL
//
// go test -bench=BenchmarkCell -benchmem

func BenchmarkCell00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = cell00(testCell)
	}
}

func BenchmarkCell01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = cell01(testCell)
	}
}

func BenchmarkCell02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = cell02(testCell)
	}
}

// HYPHEN
//
// go test -bench=BenchmarkHyp -benchmem

func BenchmarkHyp00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = hyp00(testHyp)
	}
}

func BenchmarkHyp01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = hyp01(testHyp)
	}
}

// PRESUFFIX
//
// go test -bench=BenchmarkPre -benchmem

func BenchmarkPre00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testPres {
			_ = pre00(s, testPre)
		}
	}
}

func BenchmarkPre01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testPres {
			_ = pre01(s, testPre)
		}
	}
}

// TRIMTHE
//
// go test -bench=BenchmarkThe -benchmem

func BenchmarkThe00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testThes {
			_ = trimThe0(s)
		}
	}
}

func BenchmarkThe01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testThes {
			_ = trimThe1(s)
		}
	}
}

// TRIMSP
//
// go test -bench=BenchmarkSP -benchmem

func BenchmarkSP00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testSPs {
			_ = trimSP00(s)
		}
	}
}

func BenchmarkSP01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testSPs {
			_ = trimSP01(s)
		}
	}
}

func BenchmarkSP02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testSPs {
			_ = TrimSPManual(s)
		}
	}
}

// STRIPSTART
//
// go test -bench=BenchmarkStart -benchmem

func BenchmarkStart00(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testStarts {
			_ = start00(input)
		}
	}
}

func BenchmarkStart01(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testStarts {
			_ = start01(input)
		}
	}
}

func BenchmarkStart02(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testStarts {
			_ = start02(input)
		}
	}
}

// STRIPCHARS
//
// go test -bench=BenchmarkChars -benchmem

func BenchmarkChars00(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testChars {
			_ = chars00(input)
		}
	}
}

func BenchmarkChars01(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testChars {
			_ = chars01(input)
		}
	}
}

func BenchmarkChars02(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range testChars {
			_ = chars02(input)
		}
	}
}

//
// Boilerplate: Some of these funcs and methods were recommendations from Gemini,
// others, are custom, or from online sources.
//

var testAbbr = []string{"ISO", "iso", "1st", "example", "running", "BBS", "something"}

func abbr00(s string) string {
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

func abbr01(s string) string {
	switch len(s) {
	case 2:
		if isAnyFold(s, "ad", "am", "au", "bc", "cd", "eu", "fm", "fr", "fx", "hq", "id", "ii", "pc", "pm", "st", "uk", "us") {
			return strings.ToUpper(s)
		}
	case 3:
		if isAnyFold(s, "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "3d") {
			return strings.ToLower(s)
		}
		if isAnyFold(s, "abc", "amf", "asm", "bbc", "bbs", "cgi", "diz", "dox", "faq", "fbi", "ftp", "fxp", "gbc", "gif", "iii", "iso", "kgb", "mp3", "pcb", "pcp", "pda", "psx", "pwa", "rom", "rpm", "ssd", "tnt", "tsr", "ufo", "usa", "uss", "vcd", "whq", "xxx") {
			return strings.ToUpper(s)
		}
	case 4:
		if isAnyFold(s, "10th", "11th", "12th", "13th", "7of9") {
			return strings.ToLower(s)
		}
		if isAnyFold(s, "acdc", "ansi", "ussr") {
			return strings.ToUpper(s)
		}
	}
	return ""
}

func isAnyFold(s string, targets ...string) bool {
	for _, target := range targets {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}

var abbreviations2 = map[string]string{
	"1st": "1st", "2nd": "2nd", "3rd": "3rd", "4th": "4th", "5th": "5th",
	"6th": "6th", "7th": "7th", "8th": "8th", "9th": "9th", "10th": "10th",
	"11th": "11th", "12th": "12th", "13th": "13th", "7of9": "7of9",
	"3d": "3D", "abc": "ABC", "acdc": "ACDC", "ansi": "ANSI", "bbs": "BBS",
	"cd": "CD", "faq": "FAQ", "ftp": "FTP", "gif": "GIF", "iso": "ISO",
	"mp3": "MP3", "pc": "PC", "rom": "ROM", "usa": "USA",
}

func abbr02(s string) string {
	return abbreviations2[strings.ToLower(s)]
}

var acronyms = []string{"3D", "ABC", "ACDC", "ANSI", "BBS", "CD", "FAQ", "FTP", "GIF", "ISO", "MP3", "PC", "ROM", "USA"} // MUST BE SORTED

func abbr03(s string) string {
	sUpper := strings.ToUpper(s)
	if _, found := slices.BinarySearch(acronyms, sUpper); found {
		return sUpper
	}
	return ""
}

var abbreviations4 = map[string]string{
	"1st": "1st", "2nd": "2nd", "3rd": "3rd", "4th": "4th", "5th": "5th",
	"6th": "6th", "7th": "7th", "8th": "8th", "9th": "9th", "10th": "10th",
	"11th": "11th", "12th": "12th", "13th": "13th", "7of9": "7of9",
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

func abbr04(s string) string {
	if res, ok := abbreviations4[s]; ok {
		return res
	}
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return abbreviations4[strings.ToLower(s)]
		}
	}

	return ""
}

func init() {
	raw := map[string]string{
		"1st": "1st", "2nd": "2nd", "3rd": "3rd", "4th": "4th", "5th": "5th",
		"6th": "6th", "7th": "7th", "8th": "8th", "9th": "9th", "10th": "10th",
		"11th": "11th", "12th": "12th", "13th": "13th", "7of9": "7of9",
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

	abbreviations5 = make(map[string]string, len(raw)*2)
	for k, v := range raw {
		abbreviations5[k] = v
		abbreviations5[strings.ToUpper(k)] = v
	}
}

var abbreviations5 map[string]string

func abbr05(s string) string {
	return abbreviations5[s]
}

func abbr06(s string) string {
	if res, ok := abbreviations5[s]; ok {
		return res
	}
	return abbreviations5[strings.ToLower(s)]
}

func amp00(s string) string {
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

	addWhitespace := regexp.MustCompile(`(\S)\&(\S)`)
	x = addWhitespace.ReplaceAllString(x, "$1 & $2")
	return x
}

const testAmp = "  &&hello&&world&&foo&bar&&  "

var (
	reTrimDupes = regexp.MustCompile(`&{2,}`)
	reTrimEdge  = regexp.MustCompile(`^&+|&+$`)
	reAddSpace  = regexp.MustCompile(`(\S)&(\S)`)
)

func amp01(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}

	x := reTrimDupes.ReplaceAllString(s, "&")
	x = reTrimEdge.ReplaceAllString(x, "")
	return reAddSpace.ReplaceAllString(x, "$1 & $2")
}

func amp02(s string) string {
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

func cell00(s string) string {
	const space = " "

	groups := strings.Split(s, ",")
	for index, group := range groups {
		fullname := strings.ToLower(strings.TrimSpace(group))
		fullname = unamp(fullname)
		words := strings.Split(fullname, space)
		last := len(words) - 1
		for i, x := range words {
			x = untrimdot(x)
			if fix := unhyphen(x); fix != "" {
				words[i] = fix
				continue
			}
			words[i] = unfix(x, i, last)
		}
		groups[index] = strings.Join(words, space)
	}
	return strings.ToUpper(strings.Join(groups, ", "))
}

var testCell = " Defacto2  demo  group. , the x bbs"

func cell02(s string) string {
	if s == "" {
		return ""
	}

	groups := strings.Split(s, ",")
	var result strings.Builder
	result.Grow(len(s))

	for gIdx, group := range groups {
		if gIdx > 0 {
			result.WriteString(", ")
		}

		fullname := strings.TrimSpace(group)
		fullname = unamp(fullname)
		words := strings.Fields(fullname)
		last := len(words) - 1

		for i, x := range words {
			if i > 0 {
				result.WriteByte(' ')
			}

			x = untrimdot(x)
			if fix := unhyphen(x); fix != "" {
				result.WriteString(toUpperFast(fix))
				continue
			}

			z := unfix(x, i, last)
			result.WriteString(toUpperFast(z))
		}
	}

	return result.String()
}

func toUpperFast(s string) string {
	hasLower := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'a' && s[i] <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return s
	}
	return strings.ToUpper(s)
}

func cell01(s string) string {
	if s == "" {
		return ""
	}

	groups := strings.Split(s, ",")
	var result strings.Builder
	result.Grow(len(s))

	for gIdx, group := range groups {
		if gIdx > 0 {
			result.WriteString(", ")
		}

		words := strings.Fields(group)
		last := len(words) - 1
		written := 0

		for i, word := range words {
			word = untrimdot(word)

			var fixed string
			if hFix := unhyphen(word); hFix != "" {
				fixed = hFix
			} else {
				fixed = unfix(word, i, last)
			}

			if fixed == "" {
				continue
			}

			if written > 0 {
				result.WriteByte(' ')
			}
			result.WriteString(fixed)
			written++
		}
	}

	out := unamp(result.String())
	return strings.ToUpper(out)
}

const testHyp = "multi-part-hyphenated-word-test"

func hyp00(w string) string {
	const hyphen = "-"
	if !strings.Contains(w, hyphen) {
		return ""
	}
	compounds := strings.Split(w, hyphen)
	last := len(compounds) - 1
	for i, word := range compounds {
		compounds[i] = unfix(word, i, last)
	}
	return strings.Join(compounds, hyphen)
}

func hyp01(w string) string {
	if !strings.Contains(w, "-") {
		return ""
	}

	count := 0
	for i := 0; i < len(w); i++ {
		if w[i] == '-' {
			count++
		}
	}
	last := count

	var builder strings.Builder
	builder.Grow(len(w) + 4)

	start := 0
	segmentIdx := 0

	for i := 0; i <= len(w); i++ {
		if i == len(w) || w[i] == '-' {
			subWord := w[start:i]
			fixed := unfix(subWord, segmentIdx, last)

			if fixed != "" {
				builder.WriteString(fixed)
			} else {
				builder.WriteString(subWord)
			}

			if i < len(w) {
				builder.WriteByte('-')
			}

			start = i + 1
			segmentIdx++
		}
	}

	return builder.String()
}

var testPres = []string{
	"12am",
	"superdox",
	"pc-game",
	"regularword",
}
var testPre = cases.Title(language.English, cases.NoLower)

func pre00(s string, title cases.Caser) string {
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

func pre01(s string, title cases.Caser) string {
	if len(s) < 3 {
		return ""
	}
	sfx := s[len(s)-2:]
	if sfx == "am" || sfx == "AM" || sfx == "Am" ||
		sfx == "pm" || sfx == "PM" || sfx == "Pm" ||
		sfx == "ad" || sfx == "AD" || sfx == "Ad" ||
		sfx == "bc" || sfx == "BC" || sfx == "Bc" {

		prefix := s[:len(s)-2]
		if _, err := strconv.Atoi(prefix); err == nil {
			return prefix + strings.ToUpper(sfx)
		}
	}

	word := strings.ToLower(s)
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

func trimThe0(name string) string {
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

func trimThe1(name string) string {
	if len(name) < 7 {
		return name
	}

	if !strings.EqualFold(name[:4], "the ") {
		return name
	}

	if strings.HasSuffix(strings.ToUpper(name), " BBS") ||
		strings.HasSuffix(strings.ToUpper(name), " FTP") {
		return strings.TrimSpace(name[4:])
	}

	return name
}

var testThes = []string{
	"The X BBS",
	"the super cool ftp",
	"The X",
	"X BBS",
	"Short",
}

func trimSP00(s string) string {
	const spaces = `\s+`
	r := regexp.MustCompile(spaces)
	return r.ReplaceAllString(s, " ")
}

var spaceRegex = regexp.MustCompile(`\s+`)

func trimSP01(s string) string {
	return spaceRegex.ReplaceAllString(s, " ")
}

func TrimSPManual(s string) string {
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

var testSPs = []string{
	"hello              world",
	"hello world",
	"  multiple   spaces   between   words  ",
}

func start00(s string) string {
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

var latinStartRegex = regexp.MustCompile(`[A-Za-z0-9À-ÖØ-öø-ÿ]`)

func start01(s string) string {
	f := latinStartRegex.FindStringIndex(s)
	if f == nil {
		return ""
	}
	if f[0] != 0 {
		return s[f[0]:]
	}
	return s
}

func start02(s string) string {
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return s[i:]
		}
	}
	return ""
}

var testStarts = []string{
	" - [*] checkbox",
	"checkbox",
	"*** 123 test",
	"--- !!! ---",
}

var validCharsRegex = regexp.MustCompile(`[^A-Za-zÀ-ÖØ-öø-ÿ0-9\-,& ]`)

func chars01(s string) string {
	return validCharsRegex.ReplaceAllString(s, "")
}

func chars02(s string) string {
	valid := func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
		switch r {
		case '-', ',', '&', ' ':
			return true
		}
		return false
	}

	hasInvalid := false
	for _, r := range s {
		if !valid(r) {
			hasInvalid = true
			break
		}
	}

	if !hasInvalid {
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

func chars00(s string) string {
	const validChars = `[^A-Za-zÀ-ÖØ-öø-ÿ0-9\-,& ]`
	r := regexp.MustCompile(validChars)
	return r.ReplaceAllString(s, "")
}

var testChars = []string{
	"Café!",
	".~[[@]hello[@]]~.",
	"Clean Releaser Name",
	"123-ABC, & Co.",
}

// The original, unoptimized funcs used internally by some benchmarks.
// They are now local funcs and, to avoid confusion, are prefixed with "un".

func unabbreviation(s string) string {
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

func unamp(s string) string {
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

func unconnect(w string, position, last int) string {
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

func unfix(w string, position, last int) string {
	if fix := unconnect(w, position, last); fix != "" {
		return fix
	}
	if fix := unabbreviation(w); fix != "" {
		return fix
	}
	title := cases.Title(language.English, cases.NoLower)
	if fix := unpresuffix(w, title); fix != "" {
		return fix
	}
	if fix := unsequence(w, position); fix != "" {
		return fix
	}
	return title.String(w)
}

func unhyphen(w string) string {
	const hyphen = "-"
	if !strings.Contains(w, hyphen) {
		return ""
	}
	compounds := strings.Split(w, hyphen)
	last := len(compounds) - 1
	for i, word := range compounds {
		compounds[i] = unfix(word, i, last)
	}
	return strings.Join(compounds, hyphen)
}

func unpresuffix(s string, title cases.Caser) string {
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

func unsequence(w string, i int) string {
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

func untrimdot(s string) string {
	t, ok := strings.CutSuffix(s, ".")
	if ok {
		return t
	}
	return s
}
