// Package runner is used to build and minify the css and js files.
package main

// Runner is a placeholder for esbuild to build css and js files.
// To use, run `go run runner/runner.go` and it will minify the css and js files.

import (
	"fmt"
	"os"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	layoutCSS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/css/layout.css"},
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
	if len(layoutCSS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", layoutCSS.Errors)
	}
	pouetJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/pouet.js"},
		Outfile:           "./public/js/pouet.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* pouet.min.js */",
		},
	})
	if len(pouetJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", pouetJS.Errors)
	}
	readmeJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/readme.js"},
		Outfile:           "./public/js/readme.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* readme.min.js */",
		},
	})
	if len(readmeJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", readmeJS.Errors)
	}
}
