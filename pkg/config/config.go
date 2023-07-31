// Package config for system environment variable configurations for the server.
package config

import (
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

// Config options for the Defacto2 server.
type Config struct {
	// IsProduction is true when the server is running in production mode.
	IsProduction bool `env:"PRODUCTION" envDefault:"false" help:"Use the production mode to log all errors and warnings to a file"`

	// MaxProcs is the maximum number of operating system threads the program can use.
	MaxProcs uint `env:"MAXPROCS" envDefault:"0" help:"Limit the number of operating system threads the program can use"`

	// HTTPPort is the port number to be used by the HTTP server.
	HTTPPort uint `env:"PORT" envDefault:"1323" help:"The port number to be used by the HTTP server"`

	// HTTPSRedirect is true when the server should redirect all HTTP requests to HTTPS.
	HTTPSRedirect bool `env:"HTTPS_REDIRECT" envDefault:"false" help:"Redirect all HTTP requests to HTTPS"`

	// Timeout is the timeout value in seconds for the HTTP server.
	Timeout uint `env:"TIMEOUT" envDefault:"5" help:"The timeout value in seconds for the HTTP server"`

	// DownloadDir is the directory path that holds the UUID named files that are served as release downloads.
	DownloadDir string `env:"DOWNLOAD" help:"The directory path that holds the UUID named fields that are served as release downloads"`

	// NoRobots is true when the server should tell all search engines to not crawl the website pages or assets.
	NoRobots bool `env:"NOROBOTS" envDefault:"false" avoid:"true" help:"Tell all search engines to not crawl the website pages or assets"`

	// LogRequests is true when the server should log all HTTP client requests to a file, except those with 200 OK responses.
	LogRequests bool `env:"REQUESTS" envDefault:"false" avoid:"true" help:"Log all HTTP client requests to a file"`

	// LogDir is the directory path that will store the server logs.
	LogDir string `env:"LOG" avoid:"true" help:"Overwrite the directory path that will store the program logs"`
}

func (c Config) String() string {
	const (
		minwidth = 2
		tabwidth = 4
		padding  = 2
		padchar  = ' '
		flags    = 0
		h1       = "Environment variable"
		h2       = "Value"
		h3       = "Variable"
		h4       = "Value type"
		h5       = "Help"
		line     = "─"
		donotuse = 5
	)

	b := new(strings.Builder)

	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprintf(w, "\t%s\t%s\t\t\n", h1, h2)
	fmt.Fprintf(w, "\t%s\t%s\t\t\n",
		strings.Repeat(line, len(h1)), strings.Repeat(line, len(h2)))

	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		val, def := values.FieldByName(field.Name), field.Tag.Get("envDefault")
		fmt.Fprintf(w, "\t%s\t%v\t%v\t\n",
			EnvPrefix+field.Tag.Get("env"),
			val,
			match(fmt.Sprint(val), def),
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	w = tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n", h3, h4, h5)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h3)), strings.Repeat(line, len(h4)), strings.Repeat(line, len(h5)))
	for j, field := range fields {
		if !field.IsExported() {
			continue
		}
		if j == donotuse {
			fmt.Fprintf(w, "\t\t\t\t\n")
			fmt.Fprintf(w, "\t\t\t  These variables below are not recommended.\t\n")
		}
		fmt.Fprintf(w, "\t%s\t%s\t",
			field.Tag.Get("env"),
			types(field.Type),
		)
		sp := ""
		if field.Tag.Get("avoid") != "" {
			sp = " "
		}
		fmt.Fprintf(w, "%s%s%s.\n",
			avoid(field.Tag.Get("avoid")),
			sp,
			field.Tag.Get("help"),
		)
	}
	w.Flush()
	return b.String()
}

// EnvPrefix is the prefix for all server environment variables.
const EnvPrefix = "DEFACTO2_"

// avoid returns a red cross if the value is not recommended.
func avoid(x string) string {
	if x == "true" {
		c := color.New(color.FgRed, color.Bold)
		return c.Sprint("✗")
	}
	return ""
}

// match returns a green tick if the value matches the default value.
func match(x, y string) string {
	if x != y {
		c := color.New(color.FgGreen, color.Bold)
		return c.Sprint("✓")
	}
	return ""
}

// types returns the string representation of the type.
func types(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "true|false"
	case reflect.Uint:
		return "number"
	default:
		return t.String()
	}
}
