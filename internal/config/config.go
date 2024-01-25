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
	NoCrawl        bool   `env:"NO_CRAWL" help:"Tell search engines to not crawl any of website pages or assets"`
	LogRequests    bool   `env:"LOG_REQUESTS" help:"Log all HTTP and HTTPS client requests including those with 200 OK responses"`
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
	// LocalMode build ldflags is set to true.
	LocalMode bool
}

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
	h5       = "Information"
	line     = "─"
	donotuse = 7
	down     = "DownloadDir"
	logr     = "LogDir"
	prev     = "PreviewDir"
	thumb    = "ThumbnailDir"
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

// nl prints a new line to the tabwriter.
func nl(w *tabwriter.Writer) {
	fmt.Fprintf(w, "\t\t\t\t\n")
}

// dir prints the directory path to the tabwriter or a warning if the path is empty.
func dir(w *tabwriter.Writer, id, s string) {
	if s != "" {
		fmt.Fprintf(w, "\t\t\tPATH →\t%s\n", s)
		return
	}
	fmt.Fprintf(w, "\t\t\tPATH →\t%s", "[NO DIRECTORY SET]")
	switch id {
	case down:
		fmt.Fprintf(w, "\tNo downloads will be served.\n")
	case prev:
		fmt.Fprintf(w, "\tNo preview images will be shown.\n")
	case thumb:
		fmt.Fprintf(w, "\tNo thumbnails will be shown.\n")
	case logr:
		fmt.Fprintf(w, "\tLogs will be printed to this terminal.\n")
	default:
		fmt.Fprintln(w)
	}
}

// lead prints the id, name, value and help text to the tabwriter.
func lead(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	help := field.Tag.Get("help")
	fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", helper.SplitAsSpaces(id), name, val, help)
}

// path prints the file and image paths to the tabwriter.
func path(w *tabwriter.Writer, id, name string, field reflect.StructField) {
	help := field.Tag.Get("help")
	switch id {
	case down:
		help = strings.Replace(help, "UUID named files", "UUID named files\n\t\t\t\t", 1)
	case prev:
		help = strings.Replace(help, "UUID named image", "UUID named image\n\t\t\t\t", 1)
	case thumb:
		help = strings.Replace(help, "UUID named squared image", "UUID named squared image\n\t\t\t\t", 1)
	}
	fmt.Fprintf(w, "\t%s\t%s\t\t%s.\n", helper.SplitAsSpaces(id), name, help)
}

// isProd prints a warning if the production mode is disabled.
func isProd(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	lead(w, id, name, val, field)
	if val.Kind() == reflect.Bool && !val.Bool() {
		fmt.Fprintf(w, "\t\t\t\t%s\n",
			"All errors and warnings will be logged to this console.")
	}
}

// httpPort prints the HTTP port number to the tabwriter.
func httpPort(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	lead(w, id, name, val, field)
	fmt.Fprintf(w, "\t\t\t\t%s\n",
		"The typical HTTP port number is 80, while for proxies it is 8080.")
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will use the default port number 1323.")
	}
}

// httpsPort prints the HTTPS port number to the tabwriter.
func httpsPort(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	lead(w, id, name, val, field)
	fmt.Fprintf(w, "\t\t\t\t%s\n",
		"The typical HTTPS port number is 443, while for proxies it is 8443.")
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will not use HTTPS.")
	}
}

// maxProcs prints the number of CPU cores to the tabwriter.
func maxProcs(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.", id, name, 0, field.Tag.Get("help"))
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\n\t\t\t\t%s\n", "This application will use all available CPU cores.")
	}
}

// googleHead prints a header for the Google OAuth2 configurations.
func googleHead(w *tabwriter.Writer, c Config) {
	if !c.ProductionMode && c.ReadMode {
		return
	}
	nl(w)
	fmt.Fprintf(w, "\t \t \t\t──────────────────────────────────────────────────────────────────────\n")
	fmt.Fprintf(w, "\t \t \t\t  The following configurations can usually be left at their defaults\n")
	fmt.Fprintf(w, "\t \t \t\t──────────────────────────────────────────────────────────────────────")
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

	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		switch field.Name {
		case "GoogleAccounts", "LocalMode":
			continue
		default:
		}
		// mode for development and readonly which is set using the go build flags.
		if c.LocalMode || (!c.ProductionMode && c.ReadMode) {
			if accountSkip(field.Name) {
				continue
			}
		}
		if c.LocalMode && localSkip(field.Name) {
			continue
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := EnvPrefix + field.Tag.Get("env")
		if before, found := strings.CutSuffix(name, ",unset"); found {
			name = before
		}
		c.fmtField(w, id, name, val, field)
	}
	w.Flush()
	return b
}

// fmtField prints the id, name, value and help text to the tabwriter.
func (c Config) fmtField(w *tabwriter.Writer,
	id, name string,
	val reflect.Value, field reflect.StructField,
) {
	switch id {
	case "ProductionMode":
		isProd(w, id, name, val, field)
	case "HTTPPort":
		httpPort(w, id, name, val, field)
	case "HTTPSPort":
		httpsPort(w, id, name, val, field)
	case down:
		nl(w)
		path(w, id, name, field)
		dir(w, id, c.PreviewDir)
	case prev:
		nl(w)
		path(w, id, name, field)
		dir(w, id, c.PreviewDir)
	case thumb:
		nl(w)
		path(w, id, name, field)
		dir(w, id, c.ThumbnailDir)
	case logr:
		nl(w)
		path(w, id, name, field)
		dir(w, id, c.LogDir)
	case "MaxProcs":
		maxProcs(w, id, name, val, field)
		googleHead(w, c)
	default:
		nl(w)
		lead(w, id, name, val, field)
	}
}

// localSkip skips the configurations that are inaccessable in local mode.
func localSkip(name string) bool {
	switch name {
	case
		"ReadMode",
		"ProductionMode",
		"HTTPSPort",
		"HTTPSRedirect",
		"NoCrawl",
		logr,
		"MaxProcs":
		return true
	}
	return false
}

// accountSkip skips the configurations that are not used when using Google OAuth2
// is not enabled or when the server is in read-only mode.
func accountSkip(name string) bool {
	switch name {
	case
		"GoogleClientID",
		"GoogleIDs",
		"SessionKey",
		"SessionMaxAge":
		return true
	}
	return false
}

// StaticThumb returns the path to the thumbnail directory.
func StaticThumb() string {
	return "/public/image/thumb"
}

// StaticOriginal returns the path to the image directory.
func StaticOriginal() string {
	return "/public/image/original"
}
