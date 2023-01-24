package cmd

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/urfave/cli/v2"
)

func Run() error {
	if args := len(os.Args[1:]); args > 0 {
		return greet()
	}
	return nil
}

/*
https://pkg.go.dev/runtime/debug
https://shibumi.dev/posts/go-18-feature/
build	-compiler=gc
build	CGO_ENABLED=1
build	CGO_CFLAGS=
build	CGO_CPPFLAGS=
build	CGO_CXXFLAGS=
build	CGO_LDFLAGS=
build	GOARCH=amd64
build	GOOS=linux
build	GOAMD64=v1
build	vcs=git
build	vcs.revision=7e22e19e829d84170072d2459e5870876df495ed
build	vcs.time=2022-04-03T16:59:50Z
build	vcs.modified=false
*/

// Command-line arguments handler placeholder.
// TODO: https://cli.urfave.org/v2/examples/full-api-example/
func greet() error {
	app := &cli.App{
		Name:     "Defacto2 webserver",
		Version:  Commit(),
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Ben Garrett",
				Email: "contact@defacto2.net",
			},
		},
		Copyright: "(c) 2023 Defacto2 & Ben Garrett",
		HelpName:  "server",
		Usage:     "Serve the Defacto2 website",
		UsageText: "server [options]",
		ArgsUsage: "[args and such]",
	}
	app.EnableBashCompletion = true
	app.HideHelp = false
	app.HideVersion = false
	app.Suggest = true
	if err := app.Run(os.Args); err != nil {
		return err
	}
	//fmt.Println(debug.ReadBuildInfo())
	os.Exit(0) // TODO: make an error to handle
	return nil
}

func Commit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "n/a"
}
