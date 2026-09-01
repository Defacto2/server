package name_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/releaser/name"
)

// These are benchmarks for various search and replace patterns.
// Benchmarks ending with 00 were the original implementations,
// while newer onces were either generated using Gemini or suggested
// using the Go standard docs and linters.

func BenchmarkHumanize00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, p := range testHumans {
			_, _ = Humanize00(p)
		}
	}
}

func BenchmarkHumanize01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, p := range testHumans {
			_, _ = Humanize01(p)
		}
	}
}

func BenchmarkHumanize02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, p := range testHumans {
			_, _ = Humanize02(p)
		}
	}
}

func BenchmarkObf00(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testObfs {
			_ = Obf00(s)
		}
	}
}

func BenchmarkObf01(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testObfs {
			_ = Obf01(s)
		}
	}
}

func BenchmarkObf02(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testObfs {
			_ = Obf02(s)
		}
	}
}

func BenchmarkObf03(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range testObfs {
			_ = Obf03(s)
		}
	}
}

var (
	testHumans = [...]name.Path{"defacto2", "elitendo", "tristar-red-sector", "INC_dox", "t&r"}
	testObfs   = [...]string{"Defacto2", "Defacto", "Elitendo", "ACiD Productions", "TDU-Jam!", "Razor 1911 Demo & Skillion"}
)

const (
	spacedAmpersand = " & " // " & " is a special case
	spacedComma     = ", "  // ", " is a special case
)

func Humanize00(p name.Path) (string, error) {
	if !name.Valid(p) {
		return "", name.ErrPath
	}
	s := strings.ToLower(string(p))
	s = strings.ReplaceAll(s, "-ampersand-", spacedAmpersand)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "*", spacedComma)
	return s, nil
}

func Humanize01(p name.Path) (string, error) {
	if !name.Valid(p) {
		return "", name.ErrPath
	}
	var builder strings.Builder
	builder.Grow(len(p) + 8)

	s := string(p)
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "-ampersand-") {
			builder.WriteString(spacedAmpersand)
			i += 11 // len("-ampersand-")
			continue
		}

		b := s[i]

		switch b {
		case '-':
			builder.WriteByte(' ')
		case '_':
			builder.WriteByte('-')
		case '*':
			builder.WriteString(spacedComma)
		default:
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			builder.WriteByte(b)
		}
		i++
	}

	return builder.String(), nil
}

var pathReplacer = strings.NewReplacer(
	"-ampersand-", spacedAmpersand,
	"-", " ",
	"_", "-",
	"*", spacedComma,
)

func Humanize02(p name.Path) (string, error) {
	if !name.Valid(p) {
		return "", name.ErrPath
	}

	s := strings.ToLower(string(p))
	return pathReplacer.Replace(s), nil
}

func Obf00(s string) name.Path {
	s = strings.TrimSpace(strings.ToLower(s))
	re := regexp.MustCompile(`[^a-z0-9\&\-\,\ ]`)
	s = re.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, spacedAmpersand, "-ampersand-")
	s = strings.ReplaceAll(s, spacedComma, "*")
	s = strings.ReplaceAll(s, " ", "-")
	return name.Path(s)
}

var obfre = regexp.MustCompile(`[^a-z0-9\&\-\,\ ]`)

func Obf01(s string) name.Path {
	s = strings.TrimSpace(strings.ToLower(s))
	s = obfre.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, spacedAmpersand, "-ampersand-")
	s = strings.ReplaceAll(s, spacedComma, "*")
	s = strings.ReplaceAll(s, " ", "-")
	return name.Path(s)
}

func Obf02(s string) name.Path {
	idx := -1
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !((b >= 'a' && b <= 'z') ||
			(b >= '0' && b <= '9') || b == '&' || b == '-' || b == ',' || b == ' ') {
			idx = i
			break
		}
	}

	if idx == -1 {
		return name.Path(s)
	}

	buf := make([]byte, idx, len(s))
	copy(buf, s[:idx])

	for i := idx; i < len(s); i++ {
		b := s[i]
		if (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '&' || b == '-' || b == ',' || b == ' ' {
			buf = append(buf, b)
		}
	}
	return name.Path(buf)
}

func allowed(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9') ||
		b == '&' || b == '-' || b == ',' || b == ' '
}

func Obf03(s string) string {
	idx := -1
	for i := 0; i < len(s); i++ {
		if !allowed(s[i]) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return s
	}

	buf := make([]byte, idx, len(s))
	copy(buf, s[:idx])

	for i := idx; i < len(s); i++ {
		b := s[i]
		if allowed(b) {
			buf = append(buf, b)
		}
	}

	return string(buf)
}
