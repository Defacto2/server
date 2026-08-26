package releaser_test

import (
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/initialism"
)

func BenchmarkCell(b *testing.B) {
	names := listNames()
	for b.Loop() {
		for n := range slices.Values(names) {
			if s := releaser.Cell(n); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkClean(b *testing.B) {
	names := listNames()
	for b.Loop() {
		for n := range slices.Values(names) {
			if s := releaser.Clean(n); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkHumanize(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Humanize(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkIndex(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Index(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkLink(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for uri := range *ins {
			path := releaser.Index(string(uri))
			if s := releaser.Link(path); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkObfuscate(b *testing.B) {
	ins := initialism.Initialisms()
	for b.Loop() {
		for n := range *ins {
			if s := releaser.Obfuscate(string(n)); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkTitle(b *testing.B) {
	for b.Loop() {
		for uri := range *initialism.Initialisms() {
			s := releaser.Index(string(uri))
			if title := releaser.Title(s); title != "" {
				_, _ = fmt.Fprintln(io.Discard, title)
			}
		}
	}
}
