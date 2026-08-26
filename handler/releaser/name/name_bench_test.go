package name_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/Defacto2/server/handler/releaser/initialism"
	"github.com/Defacto2/server/handler/releaser/name"
)

func BenchmarkPath(b *testing.B) {
	for b.Loop() {
		for uri := range *initialism.Initialisms() {
			path := name.Path(uri)
			if !path.Valid() {
				fmt.Fprintln(os.Stderr, "invalid! "+path.String())
				continue
			}
			if s := path.String(); s != "" {
				_, _ = fmt.Fprintln(io.Discard, s)
			}
		}
	}
}

func BenchmarkObfuscate(b *testing.B) {
	for b.Loop() {
		for i, n := range listNames() {
			_, _ = fmt.Fprintln(io.Discard, i, n, string(name.Obfuscate(n)))
		}
	}
}
