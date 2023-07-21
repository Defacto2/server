// Runner is a placeholder for esbuild to build css and js files.
package main

import (
	"fmt"
	"os"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./public/css/layout.css"},
		Outfile:           "./public/css/layout.min.css",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"css": "/* layout.min.css */",
		},
	})

	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", result.Errors)
	}
}
