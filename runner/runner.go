// Package runner is used to build and minify the css and js files.
package main

// Runner is a placeholder for esbuild to build css and js files.
// To use, run `go run runner/runner.go` and it will minify the css and js files.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

const (
	C          = "© Defacto2"
	ECMAScript = api.ES2020
)

// NamedCSS returns the base filenames of the CSS files to build.
// The files are located in the assets/css directory.
func NamedCSS() []string {
	return []string{"layout"}
}

// NamedJS returns the base filenames of the JS files to build.
// The files are located in the assets/js directory.
func NamedJS() []string {
	return []string{
		"editor-forapproval",
		"htmx-response-targets",
		"votes-pouet",
	}
}

func CanvasAnsi() api.BuildOptions {
	min := "canvas-ansi.min.js"
	entryjs := filepath.Join("assets", "js", "canvas-ansi.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func CanvasReadme() api.BuildOptions {
	min := "canvas-readme.min.js"
	entryjs := filepath.Join("assets", "js", "canvas-readme.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

// CSS are the options to build the minified CSS files.
func CSS(name string) api.BuildOptions {
	min := name + ".min.css"
	entry := filepath.Join("assets", "css", name+".css")
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

// JS are the options to build the minified JS files.
func JS(name string) api.BuildOptions {
	min := name + ".min.js"
	entry := filepath.Join("assets", "js", name+".js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entry},
		Outfile:           output,
		Target:            ECMAScript,
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

func Artifact() api.BuildOptions {
	min := "editor-artifact.min.js"
	entryjs := filepath.Join("assets", "js", "editor-artifact.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func EditorAssets() api.BuildOptions {
	min := "editor-assets.min.js"
	entryjs := filepath.Join("assets", "js", "editor-assets.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func Layout() api.BuildOptions {
	min := "layout.min.js"
	entryjs := filepath.Join("assets", "js", "layout.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func HtmxExts() api.BuildOptions {
	min := "htmx-response-targets.min.js"
	entryjs := filepath.Join("assets", "js", "htmx-response-targets.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	}
}

func Uploader() api.BuildOptions {
	min := "uploader.min.js"
	entryjs := filepath.Join("assets", "js", "uploader.js")
	output := filepath.Join("public", "js", min)
	return api.BuildOptions{
		EntryPoints:       []string{entryjs},
		Outfile:           output,
		Target:            ECMAScript,
		Write:             true,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": fmt.Sprintf("/* %s %s %s */", min, C, time.Now().Format("2006")),
		},
	}
}

func main() {
	w := os.Stderr
	for name := range slices.Values(NamedCSS()) {
		result := api.Build(CSS(name))
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "CSS build failed: %v\n", result.Errors)
		}
	}
	for name := range slices.Values(NamedJS()) {
		result := api.Build(JS(name))
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(CanvasAnsi())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(CanvasReadme())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(EditorAssets())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(Artifact())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(HtmxExts())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(Layout())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
	{
		result := api.Build(Uploader())
		if len(result.Errors) > 0 {
			fmt.Fprintf(w, "JS build failed: %v\n", result.Errors)
		}
	}
}
