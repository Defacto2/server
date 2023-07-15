package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Defacto2/server/pkg/config"
	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	Author  = "Ben Garrett"              // Author is the primary programmer of this program.
	Domain  = "defacto2.net"             // Domain of the website.
	Email   = "contact@defacto2.net"     // Email contact for public display.
	Program = "df2-server"               // Program is the command line name of this program.
	Title   = "Defacto2 web application" // Title of this program.
)

// Run parses optional command line arguments for this program.
func Run(version string, c *config.Config) (int, error) {
	if args := len(os.Args[1:]); args > 0 {
		return run(version, c)
	}
	return -1, nil
}

func run(version string, c *config.Config) (int, error) {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "configs",
				Aliases: []string{"c"},
				Usage:   "List the environment variables used for configuration",
				Action: func(ctx *cli.Context) error {
					fmt.Printf("\n%s\n", c.String())
					return nil
				},
			},
		},
		Name:     Title,
		Version:  Commit(version),
		Compiled: versioninfo.LastCommit,
		Authors: []*cli.Author{
			{
				Name:  Author,
				Email: Email,
			},
		},
		Copyright: Copyright(),
		HelpName:  Program,
		Usage:     "Serve the Defacto2 website",
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

// Arch returns the program's architecture.
func Arch() string {
	switch strings.ToLower(runtime.GOARCH) {
	case "amd64":
		return "Intel/AMD 64"
	case "arm":
		return "ARM 32"
	case "arm64":
		return "ARM 64"
	case "i386":
		return "x86"
	case "wasm":
		return "WebAssembly"
	}
	return runtime.GOARCH
}

// Commit returns a formatted, git commit description for this repository,
// including tag version and date.
func Commit(version string) string {
	s := ""
	c := versioninfo.Short()

	if version != "" {
		s = fmt.Sprintf("%s ", Vers(version))
	} else if c != "" {
		s += fmt.Sprintf("version %s, ", c)
	}
	if l := LastCommit(); l != "" {
		s += fmt.Sprintf("built in %s,", l)
	}
	if s == "" {
		return "n/a"
	}
	return strings.TrimSpace(s)
}

// Copyright returns the copyright years and author of this program.
func Copyright() string {
	const initYear = 2023
	t := versioninfo.LastCommit
	s := fmt.Sprintf("© %d Defacto2 & %s", initYear, Author)
	if t.Year() > initYear {
		s += "-" + t.Local().Format("06")
	}
	return s
}

func LastCommit() string {
	d := versioninfo.LastCommit
	if d.IsZero() {
		return ""
	}
	return d.Local().Format("2006 Jan 2 15:04")
}

// OS returns this program's operating system.
func OS() string {
	t := cases.Title(language.English)
	os := strings.Split(runtime.GOOS, "/")[0]
	switch os {
	case "darwin":
		return "macOS"
	case "freebsd":
		return "FreeBSD"
	case "js":
		return "JS"
	case "netbsd":
		return "NetBSD"
	case "openbsd":
		return "OpenBSD"
	}
	return t.String(os)
}

// Vers returns a formatted version.
// The version string is generated by GoReleaser.
func Vers(version string) string {
	const alpha, beta = "\u03b1", "β"
	if version == "" {
		return fmt.Sprintf("v0.0.0 %slpha", alpha)
	}
	const next = "-next"
	if strings.HasSuffix(version, next) {
		return fmt.Sprintf("v%s %seta", strings.TrimSuffix(version, next), beta)
	}
	return version
}
