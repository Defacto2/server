// Package cmd provides the command line interface for the Defacto2 server application.
// With the configuration of the application done using the environment variables,
// the use of commands should be kept to a minimum.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/config"
	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	Title   = "Defacto2 web application" // Title of this program.
	Domain  = "defacto2.net"             // Domain of the website.
	Program = "df2-server"               // Program is the command line name of this program.
	Author  = "Ben Garrett"              // Author is the primary programmer of this program.
	Email   = "contact@defacto2.net"     // Email contact for public display.
)

var ErrCmd = errors.New("cannot run command as config is nil")

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
		Description: desc(c),
		Compiled:    versioninfo.LastCommit,
		Copyright:   Copyright(),
		HelpName:    Program,
		Authors: []*cli.Author{
			{
				Name:  Author,
				Email: Email,
			},
		},
		Commands: []*cli.Command{
			Config(c),
			Address(c),
		},
	}
	return app
}

// Address is the `address` command help and action.
func Address(c *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "address",
		Aliases:     []string{"a"},
		Usage:       "list the server addresses",
		Description: "List the IP, hostname and port addresses the server is most probably listening on.",
		Action: func(_ *cli.Context) error {
			s, err := c.AddressesCLI()
			if err != nil {
				return fmt.Errorf("c.AddressesCLI: %w", err)
			}
			defer fmt.Fprintf(os.Stdout, "%s\n", s)
			return nil
		},
	}
}

// Config is the `config` command help and action.
func Config(c *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "list the server configuration",
		Description: "List the available server configuration options and the settings.",
		Action: func(_ *cli.Context) error {
			defer fmt.Fprintf(os.Stdout, "%s\n", c.String())
			defer func() {
				b := new(strings.Builder)
				fmt.Fprintf(os.Stdout, "%s\n", b.String())
			}()
			return nil
		},
	}
}

// Arch returns the program CPU architecture.
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

// Commit returns a formatted, git commit description for the repository,
// including tag version and date.
func Commit(ver string) string {
	x := []string{}
	s := versioninfo.Short()
	if ver != "" {
		x = append(x, Vers(ver))
	} else if s != "" {
		x = append(x, s)
	}
	built := "built at"
	if l := LastCommit(); l != "" && !strings.Contains(ver, built) {
		x = append(x, built+" "+l)
	}
	if len(x) == 0 || x[0] == "devel" {
		return "n/a (not a build)"
	}
	return strings.Join(x, ", ")
}

// Copyright returns the © symbol, years and author of this program.
// The final year is generated from the last commit date.
func Copyright() string {
	const initYear = 2023
	years := strconv.Itoa(initYear)
	t := versioninfo.LastCommit
	if t.Year() > initYear {
		years += "-" + t.Local().Format("06") //nolint:gosmopolitan
	}
	s := fmt.Sprintf("© %s Defacto2 & %s", years, Author)
	return s
}

// LastCommit returns the date of the last repository commit.
func LastCommit() string {
	d := versioninfo.LastCommit
	if d.IsZero() {
		return ""
	}
	return d.Local().Format("2006 Jan 2 15:04") //nolint:gosmopolitan
}

// OS returns the program operating system.
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
// The version string is generated by [GoReleaser].
//
// [GoReleaser]: https://goreleaser.com/
func Vers(version string) string {
	const alpha, beta = "\u03b1", "β"
	pref := Program + " version "
	if version == "" {
		return fmt.Sprintf("%s0.0.0 %slpha", pref, alpha)
	}
	const next = "-next"
	if strings.HasSuffix(version, next) {
		return fmt.Sprintf("%s%s %seta", pref, strings.TrimSuffix(version, next), beta)
	}
	return pref + version
}

// Version returns a formatted version string for this program
// including the [Commit], [OS] and CPU [Arch].
func Version(s string) string {
	x := []string{Commit(s)}
	x = append(x, fmt.Sprintf("%s on %s", OS(), Arch()))
	return strings.Join(x, " for ")
}

type ExitCode int // ExitCode is the exit code for this program.

const (
	Continue     ExitCode = iota - 1 // Continue is a special case to indicate the program should not exit.
	ExitOK                           // ExitOK is the exit code for a successful run.
	GenericError                     // GenericError is the exit code for a generic error.
	UsageError                       // UsageError is the exit code for an incorrect command line argument or usage.
)

// Run parses optional command line arguments for this program.
func Run(ver string, c *config.Config) (ExitCode, error) {
	if c == nil {
		return UsageError, ErrCmd
	}
	args := os.Args[1:]
	useArguments := len(args) > 0
	if useArguments {
		return setup(ver, c)
	}
	return Continue, nil
}

// desc returns the description for this program.
func desc(c *config.Config) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf(`Launch the web server and listen on the configured port %d.
The server expects the Defacto2 PostgreSQL database to run on the host system
or in a container. But will run without a database connection, limiting functionality.

The server relies on system environment variables for configuration and has limited 
defaults for poor usability. Without the downloads and image directories, the server 
will not display any thumbnails or previews or serve the file downloads.`, c.HTTPPort)
}

func setup(ver string, c *config.Config) (ExitCode, error) {
	if c == nil {
		return UsageError, ErrCmd
	}
	app := App(ver, c)
	app.EnableBashCompletion = true
	app.HideHelpCommand = true
	app.HideVersion = false
	app.Suggest = true
	if err := app.Run(os.Args); err != nil {
		return GenericError, fmt.Errorf("app.Run: %w", err)
	}
	return ExitOK, nil
}
