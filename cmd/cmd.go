// Package cmd provides the command line interface for the Defacto2 website application.
// These should be kept to a minimum and only used for development and debugging.
// Configuration of the web server is done via the environment variables.
package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	Program = "df2-server"               // Program is the command line name of this program.
	Domain  = "defacto2.net"             // Domain of the website.
	Title   = "Defacto2 web application" // Title of this program.
	Author  = "Ben Garrett"              // Author is the primary programmer of this program.
	Email   = "contact@defacto2.net"     // Email contact for public display.
	desc    = `Launch the server and listen on the configured port (default: 1323).
The server expects the Defacto2 PostgreSQL database running on the host system
or in a container. But will run without a database connection for debugging.`
)

var ErrCmd = fmt.Errorf("cannot run command as config is nil")

// Run parses optional command line arguments for this program.
func Run(ver string, c *config.Config) (int, error) {
	if c == nil {
		return -1, ErrCmd
	}
	if args := len(os.Args[1:]); args > 0 {
		return setup(ver, c)
	}
	return -1, nil
}

func setup(ver string, c *config.Config) (int, error) {
	if c == nil {
		return -1, ErrCmd
	}
	app := App(ver, c)
	app.EnableBashCompletion = true
	app.HideHelpCommand = true
	app.HideVersion = false
	app.Suggest = true
	if err := app.Run(os.Args); err != nil {
		return 1, err
	}
	return 0, nil
}

// App returns the command line interface for this program.
// It uses the [github.com/urfave.cli] package.
func App(ver string, c *config.Config) *cli.App {
	app := &cli.App{
		Name:    Title,
		Version: Version(ver),
		Usage:   "serve the Defacto2 web site",
		UsageText: "df2-server" +
			"\ndf2-server [command]" +
			"\ndf2-server [command] --help" +
			"\ndf2-server [flag]",
		Description: desc,
		Compiled:    versioninfo.LastCommit,
		Copyright:   Copyright(),
		HelpName:    Program,
		Authors: []*cli.Author{
			{
				Name:  Author,
				Email: Email,
			},
		},
		Commands: []*cli.Command{Config(c), Address(c)},
	}
	return app
}

// Config is the config command help and action.
func Config(c *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "list the server configuration",
		Description: "List the available server configuration options and the settings.",
		Action: func(ctx *cli.Context) error {
			defer fmt.Printf("%s\n", c.String())
			defer func() {
				x := postgres.Connection{}
				x.New()
				b := new(strings.Builder)
				x.Configurations(b)
				fmt.Printf("%s\n", b.String())
			}()
			return nil
		},
	}
}

// Address is the address command help and action.
func Address(c *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "address",
		Aliases:     []string{"a"},
		Usage:       "list the server addresses",
		Description: "List the IP, hostname and port addresses the server is most probably listening on.",
		Action: func(ctx *cli.Context) error {
			defer fmt.Printf("%s\n", c.Addresses())
			return nil
		},
	}
}

// Version returns a formatted version string for this program.
func Version(s string) string {
	x := []string{Commit(s)}
	x = append(x, fmt.Sprintf("%s for %s", OS(), Arch()))
	return strings.Join(x, " on ")
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
func Commit(ver string) string {
	x := []string{}
	s := versioninfo.Short()
	if ver != "" {
		x = append(x, Vers(ver))
	} else if s != "" {
		x = append(x, s)
	}
	if l := LastCommit(); l != "" {
		comm := fmt.Sprintf("built in %s", l)
		x = append(x, comm)
	}
	if len(x) == 0 || x[0] == "devel" {
		return "n/a (not a build)"
	}
	return strings.Join(x, ", ")
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

// LastCommit returns the date of the last repository commit.
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
