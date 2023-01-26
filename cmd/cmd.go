package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// Build information.
type Build struct {
	Version string
	Date    string
}

func (b *Build) Run() (int, error) {
	if args := len(os.Args[1:]); args > 0 {
		return b.run()
	}
	return -1, nil
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
func (b *Build) run() (int, error) {
	app := &cli.App{
		Name:     "Defacto2 webserver",
		Version:  b.Commit(),
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
		return 1, err
	}
	return 0, nil
}

func (b *Build) Commit() string {
	v, c, d := b.Version, "", b.Date
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				c = setting.Value
				break
			}
		}
	}
	s := ""
	if v != "" {
		s = fmt.Sprintf("v%s ", v)
		if d != "" {
			s += fmt.Sprintf("built on %s ", d)
		}
	} else if d != "" {
		s = fmt.Sprintf("Built on %s ", d)
	}
	if c != "" {
		s = fmt.Sprintf(" [%s]", v)
	}
	if s == "" {
		return "n/a"
	}
	return strings.TrimSpace(s)
}
