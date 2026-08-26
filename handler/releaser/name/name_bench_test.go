package name_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/handler/releaser/initialism"
	"github.com/Defacto2/server/handler/releaser/name"
)

var ins = initialism.Initialisms() //nolint:gochecknoglobals

func listNames(b *testing.B) []string {
	b.Helper()
	l := len(ins)
	n := make([]string, l)
	i := 0
	for k := range ins {
		n[i] = releaser.Humanize(string(k))
		i++
	}
	return n
}

func BenchmarkPath(b *testing.B) {
	for b.Loop() {
		for uri := range ins {
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
		for i, n := range listNames(b) {
			_, _ = fmt.Fprintln(io.Discard, i, n, string(name.Obfuscate(n)))
		}
	}
}
