package logs_test

import (
	"log/slog"
	"testing"

	"github.com/Defacto2/server/internal/logs"
)

// Benchmark ReplaceAttr with the optimization (cached ToLower).
func BenchmarkReplaceAttr(b *testing.B) {
	a := slog.String("error", "test error message")
	b.Run("", func(b *testing.B) {
		for range b.N {
			logs.ReplaceAttr(a)
		}
	})
}

// Benchmark ConfigUnsetAttr with strings.CutSuffix optimization.
func BenchmarkConfigUnsetAttr(b *testing.B) {
	a := slog.String("postgres,unset", "localhost")
	b.Run("", func(b *testing.B) {
		for range b.N {
			logs.ConfigUnsetAttr(a)
		}
	})
}

// Benchmark ConfigUnsetAttr with non-matching suffix (common case).
func BenchmarkConfigUnsetAttrNoMatch(b *testing.B) {
	a := slog.String("postgres", "localhost")
	b.Run("", func(b *testing.B) {
		for range b.N {
			logs.ConfigUnsetAttr(a)
		}
	})
}

// Benchmark ConfigIssueAttr with the double-wrapping fix.
func BenchmarkConfigIssueAttr(b *testing.B) {
	a := slog.String("issue", "database error")
	b.Run("", func(b *testing.B) {
		for range b.N {
			logs.ConfigIssueAttr(a)
		}
	})
}

// Benchmark Files.New with pre-allocated slice.
func BenchmarkFilesNew(b *testing.B) {
	f := logs.NoFiles()
	b.Run("", func(b *testing.B) {
		for range b.N {
			f.New(logs.LevelInfo, logs.Defaults)
		}
	})
}
