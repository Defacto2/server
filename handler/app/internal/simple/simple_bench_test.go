package simple_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app/internal/simple"
)

func BenchmarkHash(b *testing.B) {
	testStrings := []string{
		"short",
		"medium length string for benchmarking",
		"This is a longer string that would be more typical of real-world usage in the application for generating stable identifiers",
		strings.Repeat("a", 100), // 100 character string
	}

	for _, str := range testStrings {
		b.Run(fmt.Sprintf("length-%d", len(str)), func(b *testing.B) {
			for range b.N {
				_ = simple.Hash(str)
			}
		})
	}
}

func BenchmarkCleanHTML(b *testing.B) {
	html := `<div class="content">
		<p class="lead">This is a <strong>test</strong> with <a href="https://example.com" class="link" id="test">links</a> and <span style="color: red;">formatting</span>.</p>
		<p>Another paragraph with &nbsp; non-breaking &amp; spaces and <data-info="test">data attributes</data-info>.</p>
	</div>`

	b.Run("", func(b *testing.B) {
		for range b.N {
			simple.CleanHTML(html)
		}
	})
}
