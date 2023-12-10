// Package config for system environment variable configurations for the Defacto2 web server.
package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/Defacto2/server/internal/helper"
)

const (
	ConfigDir = "defacto2-app" // ConfigDir is the subdirectory for the home user ".config".
	EnvPrefix = "D2_"          // EnvPrefix is the prefix for all server environment variables.
)

// Config options for the Defacto2 server.
//
//nolint:lll,tagalign // The struct fields are long and the tags cannot be aligned.
type Config struct {
	// IsProduction is true when the server is running in production mode.
	IsProduction bool `env:"PRODUCTION" envDefault:"false" help:"Use the production mode to log errors to a file and recover from panics"`

	// DownloadDir is the directory path that holds the UUID named files that are served as release downloads.
	DownloadDir string `env:"DIR_DOWN" help:"The directory path that holds the UUID named files that are served as release downloads"`

	// ScreenshotDir is the directory path that holds the UUID named image files that are served as release screenshots.
	ScreenshotDir string `env:"DIR_SHOT" help:"The directory path that holds the UUID named image files that are served as release screenshots"`

	// ThumbnailDir is the directory path that holds the UUID named squared image files that are served as release thumbnails.
	ThumbnailDir string `env:"DIR_THUMB" help:"The directory path that holds the UUID named squared image files that are served as release thumbnails"`

	// HTTPPort is the port number to be used by the HTTP server.
	HTTPPort uint `env:"PORTX" envDefault:"1323" help:"The port number to be used by the unencrypted HTTP web server"`

	// HTTPSPort is the port number to be used by the HTTPS server.
	HTTPSPort uint `env:"PORTS" envDefault:"0" help:"The port number to be used by the encrypted HTTPS web server"`

	// Timeout is the timeout value in seconds for the HTTP server.
	Timeout uint `env:"TIMEOUT" envDefault:"5" help:"The timeout value in seconds for the HTTP, HTTPS and database server requests"`

	// MaxProcs is the maximum number of operating system threads the program can use.
	MaxProcs uint `env:"MAXPROCS" envDefault:"0" avoid:"true" help:"Limit the number of operating system threads the program can use"`

	// HTTPSRedirect is true when the server should redirect all HTTP requests to HTTPS.
	HTTPSRedirect bool `env:"REDIRECT" envDefault:"false" avoid:"true" help:"Redirect all HTTP requests to HTTPS"`

	// NoRobots is true when the server should tell all search engines to not crawl the website pages or assets.
	NoRobots bool `env:"NOROBOTS" envDefault:"false" avoid:"true" help:"Tell all search engines to not crawl any of website pages or assets"`

	// LogRequests is true when the server should log all HTTP client requests to a file, except those with 200 OK responses.
	LogRequests bool `env:"REQUESTS" envDefault:"false" avoid:"true" help:"Log all HTTP and HTTPS client requests to a file"`

	// LogDir is the directory path that will store the server logs.
	LogDir string `env:"DIR_LOG" avoid:"true" help:"The directory path that will store the program logs"`
}

const (
	minwidth = 2
	tabwidth = 4
	padding  = 2
	padchar  = ' '
	flags    = 0
	h1       = "Configuration"
	h2       = "Value"
	h3       = "Env variable"
	h4       = "Value type"
	h5       = "Information"
	line     = "─"
	donotuse = 7
)

// String returns a string representation of the Config struct.
// The output is formatted as a table with the following columns:
// Environment variable and Value.
func (c Config) String() string {
	b := new(strings.Builder)
	c.configurations(b)
	fmt.Fprintf(b, "\n")
	return b.String()
}

// Addresses returns a list of urls that the server is accessible from.
func (c Config) Addresses() string {
	b := new(strings.Builder)
	c.addresses(b, true)
	return b.String()
}

// Startup returns a list of urls that the server is accessible from,
// without any CLI helper text.
func (c Config) Startup() string {
	b := new(strings.Builder)
	c.addresses(b, false)
	return b.String()
}

// addresses prints a list of urls that the server is accessible from.
func (c Config) addresses(b *strings.Builder, intro bool) *strings.Builder {
	pad := strings.Repeat(string(padchar), padding)
	values := reflect.ValueOf(c)
	if intro {
		fmt.Fprintf(b, "%s\n",
			"Depending on your firewall, network and certificate setup,")
		fmt.Fprintf(b, "%s\n",
			"this web server could be accessible from the following addresses:")
		fmt.Fprintf(b, "\n")
	}
	hosts, err := helper.GetLocalHosts()
	if err != nil {
		log.Fatalf("The server cannot get the local host names: %s.", err)
	}
	port := values.FieldByName("HTTPPort").Uint()
	ports := values.FieldByName("HTTPSPort").Uint()
	const web = 80
	const webs = 443
	const local = 1323
	for _, host := range hosts {
		switch port {
		case web:
			fmt.Fprintf(b, "%shttp://%s\n", pad, host)
		case 0:
			fmt.Fprintf(b, "%shttp://%s:%d\n", pad, host, local)
		default:
			fmt.Fprintf(b, "%shttp://%s:%d\n", pad, host, port)
		}
		switch ports {
		case webs:
			fmt.Fprintf(b, "%shttps://%s\n", pad, host)
		case 0:
			// disabled
		default:
			fmt.Fprintf(b, "%shttps://%s:%d\n", pad, host, ports)
		}
	}
	ips, err := helper.GetLocalIPs()
	if err != nil {
		log.Fatalf("The server cannot get the local IP addresses: %s.", err)
	}
	for _, ip := range ips {
		fmt.Fprintf(b, "%shttp://%s:%d\n", pad, ip, port)
	}
	return b
}

// configurations prints a list of active configurations options.
func (c Config) configurations(b *strings.Builder) *strings.Builder {
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	nl := func() {
		fmt.Fprintf(w, "\t\t\t\t\n")
	}
	dir := func(s string) {
		if s != "" {
			fmt.Fprintf(w, "\t\t\tPATH →\t%s\n", s)
			return
		}
		fmt.Fprintf(w, "\t\t\tPATH →\t%s\n", "[NO DIRECTORY SET]")
	}

	fmt.Fprint(b, "Defacto2 server active configuration options.\n\n")
	fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\n",
		h1, h3, h2, h5)
	fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h1)),
		strings.Repeat(line, len(h3)),
		strings.Repeat(line, len(h2)),
		strings.Repeat(line, len(h5)))

	for j, field := range fields {
		if !field.IsExported() {
			continue
		}
		if j == donotuse {
			nl()
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := EnvPrefix + field.Tag.Get("env")
		lead := func() {
			help := field.Tag.Get("help")
			switch id {
			case "Timeout":
				help = strings.Replace(help, "HTTP, HTTPS", "HTTP, HTTPS\n\t\t\t\t", 1)
			}
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, val, help)
		}
		path := func() {
			help := field.Tag.Get("help")
			switch id {
			case "DownloadDir":
				help = strings.Replace(help, "UUID named files", "UUID named files\n\t\t\t\t", 1)
			case "ScreenshotDir":
				help = strings.Replace(help, "UUID named image", "UUID named image\n\t\t\t\t", 1)
			case "ThumbnailDir":
				help = strings.Replace(help, "UUID named squared image", "UUID named squared image\n\t\t\t\t", 1)
			}
			fmt.Fprintf(w, "\t%s\t%s\t\t%s.\n", id, name, help)
		}
		switch id {
		case "IsProduction":
			lead()
			if val.Kind() == reflect.Bool && !val.Bool() {
				fmt.Fprintf(w, "\t\t\t\t%s\n",
					"All errors and warnings will be logged to this console.")
			}
		case "HTTPPort":
			nl()
			lead()
			fmt.Fprintf(w, "\t\t\t\t%s\n",
				"The typical HTTP port number is 80, while for proxies it is 8080.")
			if val.Kind() == reflect.Uint && val.Uint() == 0 {
				fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will use the default port number 1323.")
			}
		case "HTTPSPort":
			nl()
			lead()
			fmt.Fprintf(w, "\t\t\t\t%s\n",
				"The typical HTTPS port number is 443, while for proxies it is 8443.")
			if val.Kind() == reflect.Uint && val.Uint() == 0 {
				fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will not use HTTPS.")
			}
		case "DownloadDir":
			nl()
			path()
			dir(c.DownloadDir)
		case "ScreenshotDir":
			nl()
			path()
			dir(c.ScreenshotDir)
		case "ThumbnailDir":
			nl()
			path()
			dir(c.ThumbnailDir)
		case "LogDir":
			nl()
			path()
			dir(c.LogDir)
		case "MaxProcs":
			nl()
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.", id, name, 0, field.Tag.Get("help"))
			if val.Kind() == reflect.Uint && val.Uint() == 0 {
				fmt.Fprintf(w, "\n\t\t\t\t%s\n", "This application will use all available CPU cores.")
			}
			nl()
			fmt.Fprintf(w, "\t✗\t✗\t\tThe following configurations can usually be left at their defaults.\n")
		default:
			nl()
			lead()
		}
	}
	w.Flush()
	return b
}

func StaticThumb() string {
	return "/public/image/thumb"
}

func StaticOriginal() string {
	return "/public/image/original"
}
