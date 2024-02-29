// Package runner is used to build and minify the css and js files.
package main

// Runner is a placeholder for esbuild to build css and js files.
// To use, run `go run runner/runner.go` and it will minify the css and js files.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

const C = "© Defacto2"

// NamedCSS returns the base filenames of the CSS files to build.
// The files are located in the assets/css directory.
func NamedCSS() []string {
	return []string{"layout"}
}

// NamedJS returns the base filenames of the JS files to build.
// The files are located in the assets/js directory.
func NamedJS() []string {
	return []string{"editor", "editor-assets", "editor-archive", "editor-forapproval", "pouet", "readme", "rest-zoo", "rest-pouet", "uploader"}
}

// CSS are the options to build the minified CSS files.
func CSS(name string) api.BuildOptions {
	min := fmt.Sprintf("%s.min.css", name)
	entry := filepath.Join("assets", "css", fmt.Sprintf("%s.css", name))
	output := filepath.Join("public", "css", min)
	return api.BuildOptions{
		EntryPoints:       []string{entry},
		Outfile:           output,
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"css": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

/*!
 * Bootstrap v5.3.0 (https://getbootstrap.com/)
 * Copyright 2011-2023 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
 * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
 */

/* readme.min.js */

// JS are the options to build the minified JS files.
func JS(name string) api.BuildOptions {
	min := fmt.Sprintf("%s.min.js", name)
	entry := filepath.Join("assets", "js", fmt.Sprintf("%s.js", name))
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entry},
		Outfile:           output,
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func main() {
	for _, name := range NamedCSS() {
		result := api.Build(CSS(name))
		if len(result.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "CSS build failed: %v\n", result.Errors)
		}
	}
	for _, name := range NamedJS() {
		result := api.Build(JS(name))
		if len(result.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "JS build failed: %v\n", result.Errors)
		}
	}
}
