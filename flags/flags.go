// Package flags provides the command line interface for the Defacto2 server application.
// With the configuration of the application done using the environment variables,
// the use of commands should be kept to a minimum.
package flags

import (
	"context"
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
	RecentYear = 2026                       // Most recent year of compilation for this program.

	configTitle  = "Web Server Configurations"
	addressTitle = "Web Server Addresses"
	fixesTitle   = "Database and Asset Fixes"
	format       = "%s\n"
)

var ErrNoConfig = errors.New("cannot run command as config is nil")

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

// App returns the command line interface for this program.
// It uses the [github.com/urfave.cli/v2] package.
//
// [github.com/urfave.cli/v2]: https://github.com/urfave/cli
func App(w io.Writer, ver string, c *config.Config) *cli.App {
	//nolint:exhaustruct // External library struct with many optional fields
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
			Check(w, c),
			Address(w, c),
			Fix(w, c),
		},
	}
	return app
}

// Config command lists the server configuration.
func Config(_ io.Writer, c *config.Config) *cli.Command {
	//nolint:exhaustruct // External library struct with many optional fields
	return &cli.Command{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "list the server configuration",
		Description: "List the available server configuration options and the settings.",
		Action: func(_ *cli.Context) error {
			log.Printf(format, configTitle)

			sl := stdoutput()
			c.Configurations(sl)

			return nil
		},
	}
}

// Check command checks the server configuration.
func Check(_ io.Writer, c *config.Config) *cli.Command {
	//nolint:exhaustruct // External library struct with many optional fields
	return &cli.Command{
		Name:        "check",
		Aliases:     []string{"k"},
		Usage:       "Check the server configuration",
		Description: "Check the server configuration options and the settings against common problems.",
		Action: func(_ *cli.Context) error {
			log.Printf(format, configTitle)

			sl := stdoutput()
			return c.Checks(context.Background(), sl)
		},
	}
}

// Address command lists the server addresses.
func Address(_ io.Writer, c *config.Config) *cli.Command {
	//nolint:exhaustruct // External library struct with many optional fields
	return &cli.Command{
		Name:        "address",
		Aliases:     []string{"a"},
		Usage:       "list the server addresses",
		Description: "List the IP, hostname and port addresses the server is most probably listening on.",
		Action: func(_ *cli.Context) error {
			log.Printf(format, addressTitle)

			sl := stdoutput()
			return c.Addresses(sl)
		},
	}
}

// Fix command fixes the database and assets.
func Fix(_ io.Writer, c *config.Config) *cli.Command {
	//nolint:exhaustruct // External library struct with many optional fields
	return &cli.Command{
		Name:        "fix",
		Aliases:     []string{"f"},
		Usage:       "fix the database and assets",
		Description: "Fix the database entries and file assets by running scans and checks.",
		Action: func(_ *cli.Context) error {
			nl := func() {
				_, _ = fmt.Fprintln(os.Stdout)
			}

			log.Printf(format, configTitle)

			cl := stdoutput()
			slog.SetDefault(cl)
			d := time.Now()
			c.Configurations(cl)
			nl()
			log.Println(fixesTitle)
			sl := logs.Default()
			slog.SetDefault(sl)
			if err := c.Fixer(context.Background(), sl, d); err != nil {
				const format = "fix command: %w"
				return fmt.Errorf(format, err)
			}
			return nil
		},
	}
}

func stdoutput() *slog.Logger {
	lf := logs.NoFiles()
	sl := lf.New(logs.LevelInfo, logs.Flags)
	slog.SetDefault(sl)
	return sl
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

// Copyright returns a "©" copyright symbol, the respective years and author of this program.
//
// The most recent copyright year is generated from the last commit date.
func Copyright() string {
	const (
		epoch   = 2023
		century = 100
	)
	years := strconv.Itoa(epoch)
	if RecentYear > epoch {
		endDigits := RecentYear % century
		years += "-" + fmt.Sprintf("%02d", endDigits)
	}

	return "© " + years + " Defacto2 & " + Author
}

// OS returns the host operating system.
func OS() string {
	titlize := cases.Title(language.English)
	system := strings.Split(runtime.GOOS, "/")
	if len(system) == 0 {
		return titlize.String(runtime.GOOS)
	}

	os := system[0]
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

	return titlize.String(os)
}

// VersionRelease returns a formatted release version.
// The version string is generated by [GoReleaser].
//
// [GoReleaser]: https://goreleaser.com/
func VersionRelease(version string) string {
	const (
		alphaSym = "\u03b1"
		betaSym  = "β"
	)

	prefix := Program + " version "
	if version == "" {
		return prefix + "0.0.0 " + alphaSym + "lpha"
	}

	const next = "-next"
	if before, found := strings.CutSuffix(version, next); found {
		return prefix + before + " " + betaSym + "eta"
	}

	return prefix + version
}

// VersionCommit returns a formatted, git commit description for the repository,
// including git tag version and git commit date.
func VersionCommit(ver string) string {
	const msg = "n/a (not a build)"

	elems := []string{}
	s := versioninfo.Short()
	if ver != "" {
		elems = append(elems, VersionRelease(ver))
	} else if s != "" {
		elems = append(elems, s)
	}
	if len(elems) == 0 || elems[0] == "devel" {
		return msg
	}

	return strings.Join(elems, ", ")
}

// Version returns a formatted version string for this program
// including the [VersionCommit], [OS] and CPU [Arch].
func Version(s string) string {
	const size = 2
	elems := make([]string, 0, size)
	elems = append(elems, VersionCommit(s))
	elems = append(elems, OS()+" on "+Arch())
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

	return setup(w, ver, c)
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
	const format = "application setup and run: %w"
	if err := app.Run(os.Args); err != nil {
		return GenericErr, fmt.Errorf(format, err)
	}
	return ExitOK, nil
}
