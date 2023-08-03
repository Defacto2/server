// Package config for system environment variable configurations for the server.
package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/Defacto2/server/pkg/helpers"
)

// EnvPrefix is the prefix for all server environment variables.
const EnvPrefix = "DEFACTO2_"

// Config options for the Defacto2 server.
//
//nolint:lll,tagalign // The struct fields are long and the tags cannot be aligned.
type Config struct {
	// IsProduction is true when the server is running in production mode.
	IsProduction bool `env:"PRODUCTION" envDefault:"false" help:"Use the production mode to log all errors and warnings to a file"`

	// HTTPPort is the port number to be used by the HTTP server.
	HTTPPort uint `env:"PORT" envDefault:"1323" help:"The port number to be used by the HTTP server"`

	// Timeout is the timeout value in seconds for the HTTP server.
	Timeout uint `env:"TIMEOUT" envDefault:"5" help:"The timeout value in seconds for the HTTP, HTTPS and database server requests"`

	// DownloadDir is the directory path that holds the UUID named files that are served as release downloads.
	DownloadDir string `env:"DOWNLOAD" help:"The directory path that holds the UUID named files that are served as release downloads"`

	// MaxProcs is the maximum number of operating system threads the program can use.
	MaxProcs uint `env:"MAXPROCS" envDefault:"0" avoid:"true" help:"Limit the number of operating system threads the program can use"`

	// HTTPSRedirect is true when the server should redirect all HTTP requests to HTTPS.
	HTTPSRedirect bool `env:"HTTPS_REDIRECT" envDefault:"false" avoid:"true" help:"Redirect all HTTP requests to HTTPS"`

	// NoRobots is true when the server should tell all search engines to not crawl the website pages or assets.
	NoRobots bool `env:"NOROBOTS" envDefault:"false" avoid:"true" help:"Tell all search engines to not crawl any of website pages or assets"`

	// LogRequests is true when the server should log all HTTP client requests to a file, except those with 200 OK responses.
	LogRequests bool `env:"REQUESTS" envDefault:"false" avoid:"true" help:"Log all HTTP and HTTPS client requests to a file"`

	// LogDir is the directory path that will store the server logs.
	LogDir string `env:"LOG" avoid:"true" help:"The directory path that will store the program logs"`
}

// String returns a string representation of the Config struct.
// The output is formatted as a table with the following columns:
// Environment variable and Value
func (c Config) String() string {
	const (
		minwidth = 2
		tabwidth = 4
		padding  = 2
		padchar  = ' '
		flags    = 0
		h1       = "Configuration"
		h2       = "Value"
		h3       = "Environment variable"
		h4       = "Value type"
		h5       = "About"
		line     = "─"
		donotuse = 4
	)

	if err := c.LogStorage(); err != nil {
		log.Fatalf("The server cannot save any logs: %s.", err)
	}

	b := new(strings.Builder)

	fmt.Fprint(b, "Defacto2 server active configuration options.\n\n")
	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		h1, h2, h5)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h1)), strings.Repeat(line, len(h2)), strings.Repeat(line, len(h5)))

	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	nl := func() {
		fmt.Fprintf(w, "\t\t\t\t\n")
	}

	for j, field := range fields {
		if !field.IsExported() {
			continue
		}
		if j == donotuse {
			nl()
		}
		val := values.FieldByName(field.Name)
		fmtVal := fmt.Sprint(val)
		if val.Kind() == reflect.String && val.String() == "" {
			fmtVal = `""`
		}
		id := field.Name
		lead := func() {
			fmt.Fprintf(w, "\t%s\t%v\t%s.\n",
				id,
				fmtVal,
				field.Tag.Get("help"),
			)
		}

		switch id {
		case "IsProduction":
			lead()
			if val.Kind() == reflect.Bool && !val.Bool() {
				fmt.Fprintf(w, "\t\t\t%s\n\n",
					"All errors and warnings will be logged to this console.")
			}
		case "HTTPPort":
			lead()
			fmt.Fprintf(w, "\t\t\t%s\n",
				"The typical HTTP port number is 80, while for proxies it is 8080.")
			if val.Kind() == reflect.Uint && val.Uint() == 0 {
				fmt.Fprintf(w, "\t\t\t%s\n\n", "The server will use the default port number 1323.")
			}
			nl()
			fmt.Fprintf(w, "\t\t\t%s\n",
				"Depending on your firewall and network setup,")
			fmt.Fprintf(w, "\t\t\t%s\n",
				"the server will be accessible from the following addresses:")
			hosts, err := helpers.GetLocalHosts()
			if err != nil {
				log.Fatalf("The server cannot get the local host names: %s.", err)
			}
			for _, host := range hosts {
				fmt.Fprintf(w, "\t\t\thttp://%s:%d\n", host, val.Uint())
			}
			ips, err := helpers.GetLocalIPs()
			if err != nil {
				log.Fatalf("The server cannot get the local IP addresses: %s.", err)
			}
			for _, ip := range ips {
				fmt.Fprintf(w, "\t\t\thttp://%s:%d\n", ip, val.Uint())
			}
			nl()

		case "LogDir":
			nl()
			fmt.Fprintf(w, "\t%s\t%v\t%s.",
				id,
				"",
				field.Tag.Get("help"),
			)
			fmt.Fprintf(w, "\n\t\t\t%s\n", c.LogDir)
		case "MaxProcs":
			nl()
			fmt.Fprintf(w, "\t%s\t%v\t%s.",
				id,
				0,
				field.Tag.Get("help"),
			)
			if val.Kind() == reflect.Uint && val.Uint() == 0 {
				fmt.Fprintf(w, "\n\t\t\t%s\n", "This application will use all available CPU cores.")
			}
		case "DownloadDir":
			nl()
			fmt.Fprintf(w, "\t%s\t%v\t%s.",
				id,
				"",
				field.Tag.Get("help"),
			)
			nl()
		default:
			fmt.Fprintf(w, "\t%s\t%v\t%s.\n",
				id,
				fmtVal,
				field.Tag.Get("help"),
			)
		}
	}
	fmt.Fprintln(w)
	w.Flush()

	fmt.Fprint(b, "The following environment variables can be used to override the active configuration options.\n\n")
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
		}
		fmt.Fprintf(w, "\t%s%s\t%s\t",
			avoid(field.Tag.Get("avoid")),
			EnvPrefix+field.Tag.Get("env"),
			types(field.Type),
		)
		fmt.Fprintf(w, "%s.\n", field.Tag.Get("help"))
	}
	w.Flush()
	fmt.Fprintf(b, "\n  ✗ The marked variables are not recommended for most situations.\n")

	return b.String()
}

// avoid returns a red cross if the value is not recommended.
func avoid(x string) string {
	if x == "true" {
		return "✗ "
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
