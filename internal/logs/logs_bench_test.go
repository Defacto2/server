package logs

import (
	"log/slog"
	"testing"
)

// Benchmark replaceAttr with the optimization (cached ToLower)
func BenchmarkReplaceAttr(b *testing.B) {
	a := slog.String("error", "test error message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replaceAttr(a)
	}
}

// Benchmark configUnsetAttr with strings.CutSuffix optimization
func BenchmarkConfigUnsetAttr(b *testing.B) {
	a := slog.String("postgres,unset", "localhost")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configUnsetAttr(a)
	}
}

// Benchmark configUnsetAttr with non-matching suffix (common case)
func BenchmarkConfigUnsetAttrNoMatch(b *testing.B) {
	a := slog.String("postgres", "localhost")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configUnsetAttr(a)
	}
}

// Benchmark configIssueAttr with the double-wrapping fix
func BenchmarkConfigIssueAttr(b *testing.B) {
	a := slog.String("issue", "database error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configIssueAttr(a)
	}
}

// Benchmark Files.New with pre-allocated slice
func BenchmarkFilesNew(b *testing.B) {
	f := NoFiles()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.New(LevelInfo, Defaults)
	}
}
