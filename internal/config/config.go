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

// Config options for the Defacto2 server using the [caarlos0/env] package.
//
// [caarlos0/env]:https://github.com/caarlos0/env
type Config struct {
	ProductionMode bool   `env:"PRODUCTION_MODE" help:"Use the production mode to log errors to a file and recover from panics"`
	ReadMode       bool   `env:"READ_ONLY" envDefault:"true" help:"Use the read-only mode to disable all POST, PUT and DELETE requests and any related user interface"`
	HTTPSRedirect  bool   `env:"HTTPS_REDIRECT" help:"Redirect all HTTP requests to HTTPS"`
	NoRobots       bool   `env:"NOROBOTS" help:"Tell all search engines to not crawl any of website pages or assets"`
	LogRequests    bool   `env:"LOG_REQUESTS" help:"Log every HTTP and HTTPS client requests to a file except those with 200 OK responses"`
	LogDir         string `env:"LOG_DIR" help:"The directory path that will store the program logs"`
	DownloadDir    string `env:"DOWNLOAD_DIR" help:"The directory path that holds the UUID named files that are served as artifact downloads"`
	PreviewDir     string `env:"PREVIEW_DIR" help:"The directory path that holds the UUID named image files that are served as previews of the artifact"`
	ThumbnailDir   string `env:"THUMBNAIL_DIR" help:"The directory path that holds the UUID named squared image files that are served as artifact thumbnails"`
	HTTPPort       uint   `env:"HTTP_PORT" envDefault:"1323" help:"The port number to be used by the unencrypted HTTP web server"`
	HTTPSPort      uint   `env:"HTTPS_PORT" help:"The port number to be used by the encrypted HTTPS web server"`
	MaxProcs       uint   `env:"MAX_PROCS" help:"Limit the number of operating system threads the program can use"`
	SessionKey     string `env:"SESSION_KEY,unset" help:"The session key for the cookie store or leave blank to generate a random key"`
	SessionMaxAge  int    `env:"SESSION_MAX_AGE" envDefault:"3" help:"The maximum age in hours for the session cookie"`
	GoogleClientID string `env:"GOOGLE_CLIENT_ID" help:"The Google OAuth2 client ID"`
	GoogleIDs      string `env:"GOOGLE_IDS,unset" help:"The Google OAuth2 accounts that are allowed to login"`

	// GoogleAccounts is a slice of Google OAuth2 accounts that are allowed to login.
	// Each account is a 48 byte slice of bytes that represents the SHA-384 hash of the unique Google ID.
	GoogleAccounts [][48]byte
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
func (c Config) addresses(b *strings.Builder, intro bool) {
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
	if port == 0 && ports == 0 {
		log.Fatalln("The server cannot start without a HTTP or a HTTPS port.")
	}
	const web = 80
	const webs = 443
	for _, host := range hosts {
		switch port {
		case web:
			fmt.Fprintf(b, "%shttp://%s\n", pad, host)
		case 0:
			// disabled
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
		if port == 0 {
			break
		}
		fmt.Fprintf(b, "%shttp://%s:%d\n", pad, ip, port)
	}
}

func nl(w *tabwriter.Writer) {
	fmt.Fprintf(w, "\t\t\t\t\n")
}

func dir(w *tabwriter.Writer, s string) {
	if s != "" {
		fmt.Fprintf(w, "\t\t\tPATH →\t%s\n", s)
		return
	}
	fmt.Fprintf(w, "\t\t\tPATH →\t%s\n", "[NO DIRECTORY SET]")
}

func lead(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	help := field.Tag.Get("help")
	fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, val, help)
}

func path(w *tabwriter.Writer, id, name string, field reflect.StructField) {
	help := field.Tag.Get("help")
	switch id {
	case "DownloadDir":
		help = strings.Replace(help, "UUID named files", "UUID named files\n\t\t\t\t", 1)
	case "PreviewDir":
		help = strings.Replace(help, "UUID named image", "UUID named image\n\t\t\t\t", 1)
	case "ThumbnailDir":
		help = strings.Replace(help, "UUID named squared image", "UUID named squared image\n\t\t\t\t", 1)
	}
	fmt.Fprintf(w, "\t%s\t%s\t\t%s.\n", id, name, help)
}

func isProd(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	lead(w, id, name, val, field)
	if val.Kind() == reflect.Bool && !val.Bool() {
		fmt.Fprintf(w, "\t\t\t\t%s\n",
			"All errors and warnings will be logged to this console.")
	}
}

func httpPort(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	lead(w, id, name, val, field)
	fmt.Fprintf(w, "\t\t\t\t%s\n",
		"The typical HTTP port number is 80, while for proxies it is 8080.")
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will use the default port number 1323.")
	}
}

func httpsPort(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	lead(w, id, name, val, field)
	fmt.Fprintf(w, "\t\t\t\t%s\n",
		"The typical HTTPS port number is 443, while for proxies it is 8443.")
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will not use HTTPS.")
	}
}

func maxProcs(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.", id, name, 0, field.Tag.Get("help"))
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\n\t\t\t\t%s\n", "This application will use all available CPU cores.")
	}
	nl(w)
	fmt.Fprintf(w, "\t \t \t\tThe following configurations can usually be left at their defaults\n")
	fmt.Fprintf(w, "\t \t \t\t──────────────────────────────────────────────────────────────────")
}

// configurations prints a list of active configurations options.
func (c Config) configurations(b *strings.Builder) *strings.Builder {
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
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
			nl(w)
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := EnvPrefix + field.Tag.Get("env")
		switch id {
		case "ProductionMode":
			isProd(w, id, name, val, field)
		case "HTTPPort":
			httpPort(w, id, name, val, field)
		case "HTTPSPort":
			httpsPort(w, id, name, val, field)
		case "DownloadDir":
			nl(w)
			path(w, id, name, field)
		case "PreviewDir":
			nl(w)
			path(w, id, name, field)
			dir(w, c.PreviewDir)
		case "ThumbnailDir":
			nl(w)
			path(w, id, name, field)
			dir(w, c.ThumbnailDir)
		case "LogDir":
			nl(w)
			path(w, id, name, field)
			dir(w, c.LogDir)
		case "MaxProcs":
			maxProcs(w, id, name, val, field)
		default:
			nl(w)
			lead(w, id, name, val, field)
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
