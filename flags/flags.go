// Package flags provides the command line interface for the Defacto2 server application.
// With the configuration of the application done using the environment variables,
// the use of commands should be kept to a minimum.
package flags

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/logs"
	"github.com/carlmjohnson/versioninfo"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	Title      = "Defacto2 web application" // Title of this program.
	Domain     = "defacto2.net"             // Domain of the website.
	Program    = "defacto2-server"          // Program is the command line name of this program.
	Author     = "Ben Garrett"              // Author is the primary programmer of this program.
	Email      = "contact@defacto2.net"     // Email contact for public display.
	RecentYear = 2025                       // Most recent year of compilation for this program.

	wsmsg = "Web Server Configurations"
	fxmsg = "Database and Asset Fixes"
	admsg = "Web Server Addresses"
)

var ErrNoConfig = errors.New("cannot run command as config is nil")

// App returns the command line interface for this program.
// It uses the [github.com/urfave.cli/v2] package.
//
// [github.com/urfave.cli/v2]: https://github.com/urfave/cli
func App(w io.Writer, ver string, c *config.Config) *cli.App {
	app := &cli.App{
		Name:    Title,
		Version: Version(ver),
		Usage:   "serve the Defacto2 web site",
		UsageText: Program +
			"\n" + Program + " [command]" +
			"\n" + Program + " [command] --help" +
			"\n" + Program + " [flag]",
		Description: desc(c),
		Copyright:   Copyright(),
		HelpName:    Program,
		Authors: []*cli.Author{
			{
				Name:  Author,
				Email: Email,
			},
		},
		Commands: []*cli.Command{
			Config(w, c),
			Address(w, c),
			Fix(w, c),
		},
	}
	return app
}

// Fix command the database and assets.
func Fix(_ io.Writer, c *config.Config) *cli.Command {
	const msg = "fix command"
	return &cli.Command{
		Name:        "fix",
		Aliases:     []string{"f"},
		Usage:       "fix the database and assets",
		Description: "Fix the database entries and file assets by running scans and checks.",
		Action: func(_ *cli.Context) error {
			cl := stdoutput()
			sl := logs.Default()
			slog.SetDefault(cl)
			d := time.Now()
			log.Printf("%s\n", wsmsg)
			c.Print(cl)
			newline()
			slog.SetDefault(sl)
			log.Println(fxmsg)
			if err := c.Fixer(sl, d); err != nil {
				return fmt.Errorf("%s: %w", msg, err)
			}
			return nil
		},
	}
}

// Address command lists the server addresses.
func Address(_ io.Writer, c *config.Config) *cli.Command {
	const msg = "address command"
	return &cli.Command{
		Name:        "address",
		Aliases:     []string{"a"},
		Usage:       "list the server addresses",
		Description: "List the IP, hostname and port addresses the server is most probably listening on.",
		Action: func(_ *cli.Context) error {
			sl := stdoutput()
			log.Printf("%s\n", admsg)
			err := c.Addresses(sl)
			if err != nil {
				return fmt.Errorf("%s: %w", msg, err)
			}
			return nil
		},
	}
}

// Config command lists the server configuration.
func Config(_ io.Writer, c *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "list the server configuration",
		Description: "List the available server configuration options and the settings.",
		Action: func(_ *cli.Context) error {
			sl := stdoutput()
			log.Printf("%s\n", wsmsg)
			c.Print(sl)
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
// including git tag version and git commit date.
func Commit(ver string) string {
	const msg = "n/a (not a build)"
	x := []string{}
	s := versioninfo.Short()
	if ver != "" {
		x = append(x, Vers(ver))
	} else if s != "" {
		x = append(x, s)
	}
	if len(x) == 0 || x[0] == "devel" {
		return msg
	}
	return strings.Join(x, ", ")
}

// Copyright returns a "©" copyright symbol, the respective years and author of this program.
//
// The most recent copyright year is generated from the last commit date.
func Copyright() string {
	const initYear = 2023
	years := strconv.Itoa(initYear)
	if RecentYear > initYear {
		const endDigits = RecentYear % 100
		years += "-" + strconv.Itoa(endDigits)
	}
	s := fmt.Sprintf("© %s Defacto2 & %s", years, Author)
	return s
}

// OS returns the program operating system.
func OS() string {
	t := cases.Title(language.English)
	x := strings.Split(runtime.GOOS, "/")
	if len(x) == 0 {
		return t.String(runtime.GOOS)
	}
	os := x[0]
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
	if before, found := strings.CutSuffix(version, next); found {
		return fmt.Sprintf("%s%s %seta", pref, before, beta)
	}
	return pref + version
}

// Version returns a formatted version string for this program
// including the [Commit], [OS] and CPU [Arch].
func Version(s string) string {
	elems := []string{Commit(s)}
	elems = append(elems, fmt.Sprintf("%s on %s", OS(), Arch()))
	return strings.Join(elems, " for ")
}

type ExitCode int // ExitCode is the exit code for this program.

const (
	Continue   ExitCode = iota - 1 // Continue is a special case to indicate the program should not exit.
	ExitOK                         // ExitOK is the exit code for a successful run.
	GenericErr                     // GenericError represents a generic error.
	UsageErr                       // UsageError is used for incorrect arguments or usage.
)

// Run parses optional command line arguments for this program.
func Run(w io.Writer, ver string, c *config.Config) (ExitCode, error) {
	if c == nil {
		return UsageErr, ErrNoConfig
	}
	const minArgs = 2
	if len(os.Args) < minArgs {
		return Continue, nil
	}
	args := os.Args[1:]
	useArgs := len(args) > 0
	if useArgs {
		return setup(w, ver, c)
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

func newline() {
	_, _ = fmt.Fprintln(os.Stdout)
}

func stdoutput() *slog.Logger {
	lf := logs.NoFiles()
	sl := lf.New(logs.LevelInfo, logs.Flags)
	slog.SetDefault(sl)
	return sl
}

func setup(w io.Writer, ver string, c *config.Config) (ExitCode, error) {
	if c == nil {
		return UsageErr, ErrNoConfig
	}
	app := App(w, ver, c)
	app.EnableBashCompletion = true
	app.HideHelpCommand = true
	app.HideVersion = false
	app.Suggest = true
	if err := app.Run(os.Args); err != nil {
		return GenericErr, fmt.Errorf("application setup and run: %w", err)
	}
	return ExitOK, nil
}
