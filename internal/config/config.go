// Package config manages the environment variable configurations.
package config

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Defacto2/server/internal/helper"
)

const (
	ConfigDir    = "defacto2-app" // ConfigDir is the subdirectory for the home user ".config".
	HTTPPort     = 1323           // HTTPPort is the default port number for the unencrypted HTTP server.
	SessionHours = 3              // SessionHours is the default number of hours for the session cookie to remain active.
	MinimumFiles = 40000          // MinimumFiles is the minimum number of unique filenames expected in an asset subdirectory.
	hide         = "XXXXXXXX"
)

var ErrNoPort = errors.New("the server cannot start without a http or a tls port")

// Config options for the Defacto2 server using the [caarlos0/env] package.
//
// [caarlos0/env]:https://github.com/caarlos0/env
type Config struct {
	AbsLog         string `env:"D2_DIR_LOG" help:"The absolute directory path will store all logs generated by this application"`
	AbsDownload    string `env:"D2_DIR_DOWNLOAD" help:"The directory path that holds the UUID named files that are served as artifact downloads"`
	AbsPreview     string `env:"D2_DIR_PREVIEW" help:"The directory path that holds the UUID named image files that are served as previews of the artifact"`
	AbsThumbnail   string `env:"D2_DIR_THUMBNAIL" help:"The directory path that holds the UUID named squared image files that are served as artifact thumbnails"`
	DatabaseURL    string `env:"D2_DATABASE_URL,unset" help:"Provide the URL of the database to which to connect"`
	SessionKey     string `env:"D2_SESSION_KEY,unset" help:"Use a fixed session key for the cookie store, which can be left blank to generate a random key"`
	GoogleClientID string `env:"D2_GOOGLE_CLIENT_ID,unset" help:"The Google OAuth2 client ID"`
	GoogleIDs      string `env:"D2_GOOGLE_IDS,unset" help:"Create a comma-separated list of Google account IDs to permit access to the editor mode"`
	MatchHost      string `env:"D2_MATCH_HOST" help:"Limits connections to the specific host or domain name; leave blank to permit connections from anywhere"`
	TLSCert        string `env:"D2_TLS_CERT" help:"An absolute file path to the TLS certificate, or leave blank to use a self-signed, localhost certificate"`
	TLSKey         string `env:"D2_TLS_KEY" help:"An absolute file path to the TLS key, or leave blank to use a self-signed, localhost key"`
	HTTPPort       uint   `env:"D2_HTTP_PORT" help:"The port number to be used by the unencrypted HTTP web server"`
	MaxProcs       uint   `env:"D2_MAX_PROCS" help:"Limit the number of operating system threads the program can use"`
	SessionMaxAge  int    `env:"D2_SESSION_MAX_AGE" help:"List the maximum number of hours for the session cookie to remain active before expiring and requiring a new login"`
	TLSPort        uint   `env:"D2_TLS_PORT" help:"The port number to be used by the encrypted, HTTPS web server"`
	Compression    bool   `env:"D2_COMPRESSION" help:"Enable gzip compression of the HTTP/HTTPS responses; you may turn this off when using a reverse proxy"`
	ProdMode       bool   `env:"D2_PROD_MODE" help:"Use the production mode to log errors to a file and recover from panics"`
	ReadOnly       bool   `env:"D2_READ_ONLY" help:"Use the read-only mode to turn off all POST, PUT, and DELETE requests and any related user interface"`
	NoCrawl        bool   `env:"D2_NO_CRAWL" help:"Tell search engines to not crawl any of website pages or assets"`
	LogAll         bool   `env:"D2_LOG_ALL" help:"Log all HTTP and HTTPS client requests including those with 200 OK responses"`
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
	h3       = "Environment variable"
	line     = "─"
	down     = "AbsDownload"
	logger   = "AbsLog"
	prev     = "AbsPreview"
	thumb    = "AbsThumbnail"
)

// String returns a string representation of the Config struct.
// The output is formatted as a table with the following columns:
// Environment variable and Value.
func (c Config) String() string {
	b := new(strings.Builder)
	c.configurations(b)
	return b.String()
}

// AddressesCLI returns a list of urls that the server is accessible from.
func (c Config) AddressesCLI() (string, error) {
	b := new(strings.Builder)
	if err := c.addresses(b, true); err != nil {
		return "", fmt.Errorf("c.addresses: %w", err)
	}
	return b.String(), nil
}

// Addresses returns a list of urls that the server is accessible from,
// without any CLI helper text.
func (c Config) Addresses() (string, error) {
	b := new(strings.Builder)
	if err := c.addresses(b, false); err != nil {
		return "", fmt.Errorf("c.addresses: %w", err)
	}
	return b.String(), nil
}

// addresses prints a list of urls that the server is accessible from.
func (c Config) addresses(b *strings.Builder, intro bool) error {
	pad := strings.Repeat(string(padchar), padding)
	values := reflect.ValueOf(c)
	addrIntro(b, intro)
	hosts, err := helper.LocalHosts()
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
		if c.MatchHost != "" && host != c.MatchHost {
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
		switch tls {
		case secure:
			fmt.Fprintf(b, "%shttps://%s\n", pad, host)
		case disable:
			continue
		default:
			fmt.Fprintf(b, "%shttps://%s:%d\n", pad, host, tls)
		}
	}
	if c.MatchHost == "" {
		return localIPs(b, port, pad)
	}
	return nil
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
	ips, err := helper.LocalIPs()
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

// dir prints the directory path to the tabwriter or a warning if the path is empty.
func dir(w *tabwriter.Writer, id, name, val string) {
	fmt.Fprintf(w, "\t%s\t%s", fmtID(id), name)
	if val != "" {
		// todo: stat the directory
		fmt.Fprintf(w, "\t%s\n", val)
		return
	}
	switch id {
	case down:
		fmt.Fprintf(w, "\tEmpty, no downloads will be served\n")
	case prev:
		fmt.Fprintf(w, "\tEmpty, no preview images will be shown\n")
	case thumb:
		fmt.Fprintf(w, "\tEmpty, no thumbnails will be shown\n")
	case logger:
		fmt.Fprintf(w, "\tEmpty, logs print to the terminal (stdout)\n")
	default:
		fmt.Fprintln(w)
	}
}

func fmtID(id string) string {
	switch id {
	case down:
		return "Downloads, directory path"
	case prev:
		return "Previews, directory path"
	case thumb:
		return "Thumbnails, directory path"
	case logger:
		return "Logs, directory path"
	case "Compression":
		return "Gzip compression"
	case "DatabaseURL":
		return "Database connection, URL"
	case "GoogleClientID":
		return "Google OAuth2 client ID"
	case "GoogleIDs":
		return "Google IDs for sign-in"
	case "LogAll":
		return "Log all HTTP requests"
	case "MaxProcs":
		return "Maximum CPU processes"
	case "MatchHost":
		return "Match hostname, domain or IP address"
	case "NoCrawl":
		return "Disallow search engine crawling"
	case "ProdMode":
		return "Production mode"
	case "ReadOnly":
		return "Read-only mode"
	case "SessionKey":
		return "Session encryption key"
	case "SessionMaxAge":
		return "Session, maximum age"
	case "TLSCert":
		return "TLS certificate, file path"
	case "TLSHost":
		return "TLS hostname"
	case "TLSKey":
		return "TLS key, file path"
	default:
		return helper.SplitAsSpaces(id)
	}
}

// value prints the id, name, value and help text to the tabwriter.
func value(w *tabwriter.Writer, id, name string, val reflect.Value) {
	if val.Kind() == reflect.Bool {
		status := "Off"
		if val.Bool() {
			status = "On"
		}
		fmt.Fprintf(w, "\t%s\t%s\t%v\n", fmtID(id), name, status)
		return
	}
	fmt.Fprintf(w, "\t%s\t%s\t", fmtID(id), name)
	switch id {
	case "GoogleClientID":
		if val.String() == "" {
			fmt.Fprint(w, "Empty, no account sign-in for web administration\n")
			return
		}
		fmt.Fprint(w, val.String())
	case "MatchHost":
		if val.String() == "" {
			fmt.Fprint(w, "Empty, no address restrictions\n")
			return
		}
		fmt.Fprint(w, val.String())
	case "SessionKey":
		if val.String() == "" {
			fmt.Fprint(w, "Empty, a random key will be generated during the server start\n")
			return
		}
		fmt.Fprint(w, hide)
	case "SessionMaxAge":
		fmt.Fprintf(w, "%v hours\n", val.Int())
	case "DatabaseURL":
		fmt.Fprint(w, hidePassword(val.String()))
	default:
		if val.String() == "" {
			fmt.Fprint(w, "Empty\n")
			return
		}
		fmt.Fprintf(w, "%v\n", val)
	}
}

// httpPort prints the HTTP port number to the tabwriter.
func httpPort(w *tabwriter.Writer, id, name string, val reflect.Value) {
	fmt.Fprintf(w, "\t%s\t%s\t", fmtID(id), name)
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "%s\n", "0, the web server will not use HTTP")
		return
	}
	port := val.Uint()
	const common = 80
	if port == common {
		fmt.Fprintf(w, "%d, the web server will use HTTP, example: http://localhost\n", port)
		return
	}
	fmt.Fprintf(w, "%d, the web server will use HTTP, example: http://localhost:%d\n", port, port)
}

// tlsPort prints the HTTPS port number to the tabwriter.
func tlsPort(w *tabwriter.Writer, id, name string, val reflect.Value) {
	fmt.Fprintf(w, "\t%s\t%s\t", fmtID(id), name)
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "%s\n", "0, the web server will not use HTTPS")
		return
	}
	port := val.Uint()
	const common = 443
	if port == common {
		fmt.Fprintf(w, "%d, the web server will use HTTPS, example: https://localhost\n", port)
		return
	}
	fmt.Fprintf(w, "%d, the web server will use HTTPS, example: https://localhost:%d\n", port, port)
}

// maxProcs prints the number of CPU cores to the tabwriter.
func maxProcs(w *tabwriter.Writer, id, name string, val reflect.Value) {
	fmt.Fprintf(w, "\t%s\t%s\t", fmtID(id), name)
	if val.Kind() == reflect.Uint && val.Uint() == 0 {
		fmt.Fprintf(w, "%s\n", "0, the application will use all available CPU threads")
		return
	}
	fmt.Fprintf(w, "%d, the application will limit access to CPU threads\n", val.Uint())
}

// hidePassword replaces the password in the URL with XXXXXs.
func hidePassword(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	_, exists := u.User.Password()
	if !exists {
		return rawURL
	}
	u.User = url.UserPassword(u.User.Username(), hide)
	return u.String()
}

// configurations prints a list of active configurations options.
func (c Config) configurations(b *strings.Builder) *strings.Builder {
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
	values := reflect.ValueOf(c)

	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	fmt.Fprint(b, "The Defacto2 server configuration:\n\n")
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		h1, h3, h2)
	fmt.Fprintf(w, "\t%s\t%s\t%s\n",
		strings.Repeat(line, len(h1)),
		strings.Repeat(line, len(h3)),
		strings.Repeat(line, len(h2)))

	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		switch field.Name {
		case "GoogleAccounts":
			continue
		default:
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := field.Tag.Get("env")
		if before, found := strings.CutSuffix(name, ",unset"); found {
			name = before
		}
		c.fmtField(w, id, name, val)
	}
	w.Flush()
	return b
}

// fmtField prints the id, name, value and help text to the tabwriter.
func (c Config) fmtField(w *tabwriter.Writer,
	id, name string,
	val reflect.Value,
) {
	fmt.Fprintf(w, "\t\t\t\t\n")
	switch id {
	case "HTTPPort":
		httpPort(w, id, name, val)
	case "TLSPort":
		tlsPort(w, id, name, val)
	case down, prev, thumb, logger:
		dir(w, id, name, val.String())
	case "MaxProcs":
		maxProcs(w, id, name, val)
	case "GoogleIDs":
		l := len(c.GoogleAccounts)
		fmt.Fprintf(w, "\t%s\t%s\t", fmtID(id), name)
		switch l {
		case 0:
			fmt.Fprint(w, "Empty, no accounts for web administration\n")
		case 1:
			fmt.Fprint(w, "1 Google account allowed to sign-in\n")
		default:
			fmt.Fprintf(w, "%d Google accounts allowed to sign-in\n", l)
		}
	default:
		value(w, id, name, val)
	}
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
func (c *Config) Override() {
	// hash and delete any supplied google ids
	ids := strings.Split(c.GoogleIDs, ",")
	for _, id := range ids {
		sum := sha512.Sum384([]byte(id))
		c.GoogleAccounts = append(c.GoogleAccounts, sum)
	}
	c.GoogleIDs = "overwrite placeholder"
	c.GoogleIDs = "" // empty the string

	// set the default HTTP port if both ports are configured to zero
	if c.HTTPPort == 0 && c.TLSPort == 0 {
		c.HTTPPort = HTTPPort
	}
}
