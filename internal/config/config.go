// Package config manages the environment variable configurations.
package config

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
)

const (
	ConfigDir = "defacto2-app" // ConfigDir is the subdirectory for the home user ".config".
	HTTPPort  = 1323           // HTTPPort is the default port number for the unencrypted HTTP server.
)

var ErrNoPort = errors.New("the server cannot start without a http or a tls port")

// Config options for the Defacto2 server using the [caarlos0/env] package.
//
// [caarlos0/env]:https://github.com/caarlos0/env
type Config struct {
	Compression    string `env:"D2_COMPRESSION" envDefault:"gzip" help:"Enable either gzip or br compression of HTTP/HTTPS responses, you may want to disable this if using a reverse proxy"`
	LogDir         string `env:"D2_LOG_DIR" help:"The directory path that will store the program logs"`
	DownloadDir    string `env:"D2_DOWNLOAD_DIR" help:"The directory path that holds the UUID named files that are served as artifact downloads"`
	PreviewDir     string `env:"D2_PREVIEW_DIR" help:"The directory path that holds the UUID named image files that are served as previews of the artifact"`
	ThumbnailDir   string `env:"D2_THUMBNAIL_DIR" help:"The directory path that holds the UUID named squared image files that are served as artifact thumbnails"`
	SessionKey     string `env:"D2_SESSION_KEY,unset" help:"The session key for the cookie store or leave blank to generate a random key"`
	GoogleClientID string `env:"D2_GOOGLE_CLIENT_ID" help:"The Google OAuth2 client ID"`
	GoogleIDs      string `env:"D2_GOOGLE_IDS,unset" help:"The Google OAuth2 accounts that are allowed to login"`
	TLSCert        string `env:"D2_TLS_CERT" help:"The TLS certificate file path, leave blank to use the self-signed, localhost certificate"`
	TLSKey         string `env:"D2_TLS_KEY" help:"The TLS key file path, leave blank to use the self-signed, localhost key"`
	TLSHost        string `env:"D2_TLS_HOST" help:"This recommended setting, limits TSL to the specific host or domain name, leave blank to permit TLS connections from any host"`

	HostName string `env:"PS_HOST_NAME"` // this should only be used internally, instead see postgres.Connection

	// GoogleAccounts is a slice of Google OAuth2 accounts that are allowed to login.
	// Each account is a 48 byte slice of bytes that represents the SHA-384 hash of the unique Google ID.
	GoogleAccounts [][48]byte
	HTTPPort       uint `env:"D2_HTTP_PORT" envDefault:"1323" help:"The port number to be used by the unencrypted HTTP web server"`
	MaxProcs       uint `env:"D2_MAX_PROCS" help:"Limit the number of operating system threads the program can use"`
	SessionMaxAge  int  `env:"D2_SESSION_MAX_AGE" envDefault:"3" help:"The maximum age in hours for the session cookie"`
	TLSPort        uint `env:"D2_TLS_PORT" help:"The port number to be used by the encrypted, HTTPS web server"`
	ProductionMode bool `env:"D2_PRODUCTION_MODE" help:"Use the production mode to log errors to a file and recover from panics"`
	FastStart      bool `env:"D2_FAST_START" help:"Skip the database connection and file checks on server startup"`
	ReadMode       bool `env:"D2_READ_ONLY" envDefault:"true" help:"Use the read-only mode to disable all POST, PUT and DELETE requests and any related user interface"`
	NoCrawl        bool `env:"D2_NO_CRAWL" help:"Tell search engines to not crawl any of website pages or assets"`
	LogRequests    bool `env:"D2_LOG_REQUESTS" help:"Log all HTTP and HTTPS client requests including those with 200 OK responses"`
	HTTPSRedirect  bool `env:"D2_HTTPS_REDIRECT" help:"Redirect all HTTP requests to HTTPS"`

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
	logger   = "LogDir"
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

// AddressesCLI returns a list of urls that the server is accessible from.
func (c Config) AddressesCLI() (string, error) {
	b := new(strings.Builder)
	if err := c.addresses(b, true); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Addresses returns a list of urls that the server is accessible from,
// without any CLI helper text.
func (c Config) Addresses() (string, error) {
	b := new(strings.Builder)
	if err := c.addresses(b, false); err != nil {
		return "", err
	}
	return b.String(), nil
}

// addresses prints a list of urls that the server is accessible from.
func (c Config) addresses(b *strings.Builder, intro bool) error {
	pad := strings.Repeat(string(padchar), padding)
	values := reflect.ValueOf(c)
	addrIntro(b, intro)
	hosts, err := helper.GetLocalHosts()
	if err != nil {
		return fmt.Errorf("the server cannot get the local host names: %w", err)
	}
	port := values.FieldByName("HTTPPort").Uint()
	tls := values.FieldByName("TLSPort").Uint()
	if port == 0 && tls == 0 {
		return ErrNoPort
	}
	const disable, text, secure = 0, 80, 443
	for _, host := range hosts {
		if c.HostName == postgres.DockerHost && host != "localhost" {
			// skip all but localhost when running in docker
			continue
		}
		switch port {
		case text:
			fmt.Fprintf(b, "%shttp://%s\n", pad, host)
		case disable:
			continue
		default:
			fmt.Fprintf(b, "%shttp://%s:%d\n", pad, host, port)
		}
		if c.TLSHost != "" && host != c.TLSHost {
			continue
		}
		switch tls {
		case secure:
			fmt.Fprintf(b, "%shttps://%s\n", pad, host)
		case disable:
			continue
		default:
			fmt.Fprintf(b, "%shttps://%s:%d\n", pad, host, tls)
		}
	}
	if c.HostName == postgres.DockerHost {
		return nil
	}
	return localIPs(b, port, pad)
}

func addrIntro(b *strings.Builder, intro bool) {
	if !intro {
		return
	}
	fmt.Fprintf(b, "%s\n",
		"Depending on your firewall, network and certificate setup,")
	fmt.Fprintf(b, "%s\n",
		"this web server could be accessible from the following addresses:")
	fmt.Fprintf(b, "\n")
}

func localIPs(b *strings.Builder, port uint64, pad string) error {
	ips, err := helper.GetLocalIPs()
	if err != nil {
		return fmt.Errorf("the server cannot get the local IP addresses: %w", err)
	}
	for _, ip := range ips {
		if port == 0 {
			break
		}
		fmt.Fprintf(b, "%shttp://%s:%d\n", pad, ip, port)
	}
	return nil
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
	case logger:
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

// tlsPort prints the HTTPS port number to the tabwriter.
func tlsPort(w *tabwriter.Writer, id, name string, val reflect.Value, field reflect.StructField) {
	nl(w)
	lead(w, id, name, val, field)
	fmt.Fprintf(w, "\t\t\t\t%s\n",
		"The typical TLS port number is 443, while for proxies it is 8443.")
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "\t\t\t\t%s\n", "The server will not use TLS.")
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
		case "GoogleAccounts", "LocalMode", "HostName":
			continue
		default:
		}
		// mode for development and readonly which is set using the go build flags.
		if c.LocalMode || (!c.ProductionMode && c.ReadMode) {
			if AccountSkip(field.Name) {
				continue
			}
		}
		if c.LocalMode && LocalSkip(field.Name) {
			continue
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := field.Tag.Get("env")
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
	case "TLSPort":
		tlsPort(w, id, name, val, field)
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
	case logger:
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

// LocalSkip skips the configurations that are inaccessible in local mode.
func LocalSkip(name string) bool {
	switch name {
	case
		"ReadMode",
		"ProductionMode",
		"TLSPort",
		"HTTPSRedirect",
		"NoCrawl",
		logger,
		"MaxProcs":
		return true
	}
	return false
}

// AccountSkip skips the configurations that are not used when using Google OAuth2
// is not enabled or when the server is in read-only mode.
func AccountSkip(name string) bool {
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

// UseTLS returns true if the server is configured to use TLS.
func (c Config) UseTLS() bool {
	return c.TLSPort > 0 && c.TLSCert != "" || c.TLSKey != ""
}

// UseHTTP returns true if the server is configured to use HTTP.
func (c Config) UseHTTP() bool {
	return c.HTTPPort > 0
}

// UseTLSLocal returns true if the server is configured to use the local-mode.
func (c Config) UseTLSLocal() bool {
	return c.TLSPort > 0 && c.TLSCert == "" && c.TLSKey == ""
}

// Override the configuration settings fetched from the environment.
func (c *Config) Override(localMode bool) {
	// Build binary, environment variables overrides using,
	// go build -ldflags="-X 'main.LocalMode=true'"
	if localMode {
		if c.HTTPPort == 0 {
			c.HTTPPort = HTTPPort
		}
		c.LocalMode = true
		c.ProductionMode = false
		c.ReadMode = true
		c.NoCrawl = true
		c.LogDir = ""
		c.GoogleClientID = ""
		c.GoogleIDs = ""
		c.SessionKey = ""
		c.SessionMaxAge = 0
		c.TLSPort = 0
		c.TLSCert = ""
		c.TLSKey = ""
		c.HTTPSRedirect = false
		c.MaxProcs = 0
		return
	}
	// hash and delete any supplied google ids
	ids := strings.Split(c.GoogleIDs, ",")
	for _, id := range ids {
		sum := sha512.Sum384([]byte(id))
		c.GoogleAccounts = append(c.GoogleAccounts, sum)
	}
	c.GoogleIDs = "overwrite placeholder"
	c.GoogleIDs = "" // empty the string

	if c.HTTPPort == 0 && c.TLSPort == 0 {
		c.HTTPPort = HTTPPort
	}
}
