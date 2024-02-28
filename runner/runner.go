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
	editorJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/editor.js"},
		Outfile:           "./public/js/editor.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* editor.min.js */",
		},
	})
	if len(editorJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", editorJS.Errors)
	}
	editorAssetsJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/editor-assets.js"},
		Outfile:           "./public/js/editor-assets.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* editor-assets.min.js */",
		},
	})
	if len(editorAssetsJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", editorAssetsJS.Errors)
	}
	editorArchiveJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/editor-archive.js"},
		Outfile:           "./public/js/editor-archive.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* editor-archive.min.js */",
		},
	})
	if len(editorArchiveJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", editorArchiveJS.Errors)
	}
	editorForApproval := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/editor-forapproval.js"},
		Outfile:           "./public/js/editor-forapproval.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* editor-forapproval.min.js */",
		},
	})
	if len(editorForApproval.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", editorForApproval.Errors)
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
	restZooJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/rest-zoo.js"},
		Outfile:           "./public/js/rest-zoo.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* rest-zoo.min.js */",
		},
	})
	if len(restZooJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", restZooJS.Errors)
	}
	restPouetJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/rest-pouet.js"},
		Outfile:           "./public/js/rest-pouet.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* rest-pouet.min.js */",
		},
	})
	if len(restPouetJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", restPouetJS.Errors)
	}
	uploaderJS := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./assets/js/uploader.js"},
		Outfile:           "./public/js/uploader.min.js",
		Write:             true,
		Bundle:            false,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Banner: map[string]string{
			"js": "/* uploader.min.js */",
		},
	})
	if len(uploaderJS.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", uploaderJS.Errors)
	}
}
