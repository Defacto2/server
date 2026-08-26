package initialism_test

import (
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/Defacto2/server/handler/releaser/initialism"
)

const guarantee = "defacto2" // must guarantee a valid match for benchmarking

func BenchmarkIsInitialism(b *testing.B) {
	for b.Loop() {
		_, _ = fmt.Fprintln(io.Discard, initialism.IsInitialism(guarantee))
	}
}

func BenchmarkInitialism(b *testing.B) {
	for b.Loop() {
		_, _ = fmt.Fprintln(io.Discard, initialism.Initialism(guarantee))
	}
}

func BenchmarkInitialisms(b *testing.B) {
	for b.Loop() {
		for key, values := range *initialism.Initialisms() {
			for value := range slices.Values(values) {
				if value == guarantee {
					_, _ = fmt.Fprintf(io.Discard, "Found %v in %v\n", guarantee, key)
					return
				}
			}
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	for b.Loop() {
		_, _ = fmt.Fprint(io.Discard, initialism.Match(guarantee))
	}
}
