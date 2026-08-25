// Package config manages the environment variable configurations.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/nils"
)

const (
	// ConfigDir is the subdirectory for the home user ".config".
	ConfigDir = "defacto2-app"
	// MinimumFiles is the minimum number of unique filenames expected in an asset subdirectory.
	MinimumFiles = 40_000
	// SessionHours is the default number of hours for the session cookie to remain active.
	SessionHours = 3
	// StdHTTP is the standard port used for a legacy unencrypted HTTP connection.
	StdHTTP Port = 80
	// StdHTTPS is the standard port used for a HTTP web connection.
	StdHTTPS Port = 443
	// StdCustom is the default port number used by this application for an unencrypted HTTP connection.
	StdCustom = 1323
)

const (
	mask = "xxxxx" // hide is the placeholder text used to replace sensitive information
)

//nolint:gochecknoglobals
var informations = map[string]string{
	"AbsDownload":    "Downloads, directory path",
	"AbsPreview":     "Previews, directory path",
	"AbsThumbnail":   "Thumbnails, directory path",
	"AbsLog":         "Logs, directory path",
	"AbsExtra":       "Extras, directory path",
	"AbsOrphaned":    "Orphaned, directory path",
	"Compression":    "Gzip compression",
	"DatabaseURL":    "Database connection, URL",
	"GoogleClientID": "Google OAuth2 client ID",
	"GoogleIDs":      "Google IDs for sign-in",
	"LogAll":         "Log all HTTP requests",
	"MaxProcs":       "Maximum CPU processes",
	"MatchHost":      "Match hostname, domain or IP address",
	"NoCrawl":        "Disallow search engine crawling",
	"ProdMode":       "Production mode",
	"Quiet":          "Quiet mode",
	"ReadOnly":       "Read-only mode",
	"SessionKey":     "Session encryption key",
	"SessionMaxAge":  "Maximum age of a session for the web administration",
	"TLSCert":        "TLS certificate, file path",
	"TLSHost":        "TLS hostname",
	"TLSKey":         "TLS key, file path",
}

// information returns a human readable description of the named configuration identifier.
func information(name string) string {
	if desc, found := informations[name]; found {
		return desc
	}
	return helper.SplitAsSpaces(name)
}

// Config options for the Defacto2 server using the [caarlos0/env] package.
//
// [caarlos0/env]:https://github.com/caarlos0/env
type Config struct { //nolint:recvcheck
	AbsLog         Abslog     `env:"D2_DIR_LOG"`                // absolute directory path to store log files
	AbsDownload    Absdown    `env:"D2_DIR_DOWNLOAD"`           // absolute directory path to the artifact file downloads
	AbsPreview     Absprev    `env:"D2_DIR_PREVIEW"`            // absolute directory path to the preview images
	AbsThumbnail   Absthumb   `env:"D2_DIR_THUMBNAIL"`          // absolute directory path to the thumbnail images
	AbsExtra       Absextra   `env:"D2_DIR_EXTRA"`              // absolute directory path to the artifact extras
	AbsOrphaned    Absorphan  `env:"D2_DIR_ORPHANED"`           // absolute directory path to the retired artifacts
	DatabaseURL    Connection `env:"D2_DATABASE_URL"`           // url to the database
	SessionKey     Sessionkey `env:"D2_SESSION_KEY,unset"`      // key or random value used for cookie encryption
	SessionMaxAge  Hours      `env:"D2_SESSION_MAX_AGE"`        // number of hours to keep a cookie session
	GoogleClientID GoogleAuth `env:"D2_GOOGLE_CLIENT_ID,unset"` // google oauth2 clien id
	GoogleIDs      GoogleID   `env:"D2_GOOGLE_IDS,unset"`       // google accounts
	HTTPPort       PortHTTP   `env:"D2_HTTP_PORT"`              // port value for serving the website using http
	TLSPort        PortTLS    `env:"D2_TLS_PORT"`               // port value for serving the website using https
	TLSCert        Abstlscrt  `env:"D2_TLS_CERT"`               // absolute path to the tls certificate file
	TLSKey         Abstlskey  `env:"D2_TLS_KEY"`                // absolute path to the tls key file
	MatchHost      Matchhost  `env:"D2_MATCH_HOST"`             // restrict connections to specific internet hosts or ip addresses
	MaxProcs       Threads    `env:"D2_MAX_PROCS"`              // retrict the number of os threads the server can use
	Quiet          Toggle     `env:"D2_QUIET"`                  // suppress startup terminal text
	Compression    Toggle     `env:"D2_COMPRESSION"`            // enable gzip compression when serving webpages
	ProdMode       Toggle     `env:"D2_PROD_MODE"`              // production mode, save log files and always recover from panics
	ReadOnly       Toggle     `env:"D2_READ_ONLY"`              // read only mode, disables all post, put, delete http requests
	NoCrawl        Toggle     `env:"D2_NO_CRAWL"`               // always insert a http header to ask search engines not to crawl the site
	LogAll         Toggle     `env:"D2_LOG_ALL"`                // log all client requests, both errors and successful responses
	GoogleAccounts OAuth2s    // GoogleAccounts is the data store for the GoogleIDs.
}

// Help returns a mapped inventory of description texts for documentation/CLI usage.
func (c Config) Help() map[string]string {
	return map[string]string{
		"D2_DIR_LOG":          "The absolute directory path will store all logs generated by this application",
		"D2_DIR_DOWNLOAD":     "The directory path that holds the UUID named files that are served as artifact downloads",
		"D2_DIR_PREVIEW":      "The directory path that holds the UUID named image files that are served as previews of the artifact",
		"D2_DIR_THUMBNAIL":    "The directory path that holds the UUID named squared image files that are served as artifact thumbnails",
		"D2_DIR_EXTRA":        "The directory path that holds extra assets of the UUID named files that are generated by the application",
		"D2_DIR_ORPHANED":     "The directory path that holds the UUID named files that are not linked to any database records",
		"D2_DATABASE_URL":     "Provide the URL of the database to which to connect",
		"D2_SESSION_KEY":      "Use a fixed session key for the cookie store, which can be left blank to generate a random key",
		"D2_GOOGLE_CLIENT_ID": "The Google OAuth2 client ID",
		"D2_GOOGLE_IDS":       "Create a comma-separated list of Google account IDs to permit access to the editor mode",
		"D2_MATCH_HOST":       "Limits connections to the specific host or domain name; leave blank to permit connections from anywhere",
		"D2_TLS_CERT":         "An absolute file path to the TLS certificate, or leave blank to use a self-signed, localhost certificate",
		"D2_TLS_KEY":          "An absolute file path to the TLS key, or leave blank to use a self-signed, localhost key",
		"D2_HTTP_PORT":        "The port number to be used by the unencrypted HTTP web server",
		"D2_MAX_PROCS":        "Limit the number of operating system threads the program can use",
		"D2_SESSION_MAX_AGE":  "List the maximum number of hours for the session cookie to remain active before expiring and requiring a new login",
		"D2_TLS_PORT":         "The port number to be used by the encrypted, HTTPS web server",
		"D2_QUIET":            "Suppress most startup output to the terminal, intended for use with systemd or other process managers",
		"D2_COMPRESSION":      "Enable gzip compression of the HTTP/HTTPS responses; you may turn this off when using a reverse proxy",
		"D2_PROD_MODE":        "Use the production mode to log errors to files and recover from panics",
		"D2_READ_ONLY":        "Use the read-only mode to turn off all POST, PUT, and DELETE requests and any related user interface",
		"D2_NO_CRAWL":         "Tell search engines to not crawl any of website pages or assets",
		"D2_LOG_ALL":          "Log all HTTP and HTTPS client requests including those with 200 OK responses",
	}
}

// Configured returns both an about text and configured description for the named configuration.
//
// The name is the configuration field such as "AbsLog" for D2_DIR_LOG.
// The value must be result of the configuration field.
func Configured(name string, value any) string {
	s := information(name)
	if value == nil {
		return s
	}

	// special cases
	switch name {
	case "GoogleAccounts":
		return fmt.Sprintf("Google account(s), %v", value)
	case "SessionMaxAge":
		return fmt.Sprintf("%s, %v", s, value)
	case "MaxProcs":
		return fmt.Sprintf("%s %v", s, value)
	}

	// toggles
	if v, ok := value.(bool); ok {
		return fmt.Sprintf("%s is %t", s, v)
	}

	// everything else
	state := func(v reflect.Value) string {
		if v.Bool() {
			return "on"
		}
		return "off"
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Bool {
		return fmt.Sprintf("%s is %s", s, state(v))
	}

	return s
}

// Issuer describes any config field capable of reporting validation issues.
type Issuer interface {
	Issue() string
}

// Helper describes any config field capable of providing contextual guidance.
type Helper interface {
	Help() string
}

// Configurations prints the application settings to the logger.
func (c Config) Configurations(sl *slog.Logger) {
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("config printer: %w", err))
	}

	val := reflect.ValueOf(c)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		fieldName := structField.Name

		if strings.EqualFold(fieldName, "GoogleAccounts") {
			continue // skip
		}

		key := strings.ToLower(structField.Tag.Get("env"))
		fieldValue := val.Field(i).Interface()
		msg := Configured(fieldName, fieldValue) + ":"

		// check for any Issue() via interface or reflection fallback
		if issuer, ok := fieldValue.(Issuer); ok {
			if issue := issuer.Issue(); issue != "" {
				sl.Error(msg, slog.Any(key, fieldValue), slog.String("issue", issue))
				continue
			}
		}

		// special case
		if fieldName == "GoogleIDs" {
			if _, ok := fieldValue.(Helper); ok {
				c.googleIDs(sl, key)
				continue
			}
		}

		// check for any Help() via interface or reflection fallback
		if helper, ok := fieldValue.(Helper); ok {
			if tip := helper.Help(); tip != "" {
				sl.Info(msg, slog.Any(key, fieldValue), slog.String("tip", tip))
				continue
			}
		}

		// default log output
		sl.Info(msg, slog.Any(key, fieldValue))
	}
}

// Names returns a list of the field names in the Config struct.
func (c Config) Names() []string {
	t := reflect.TypeFor[Config]()
	fieldNames := make([]string, t.NumField())
	for i := range t.NumField() {
		fieldNames[i] = t.Field(i).Name
	}
	return fieldNames
}

// Addresses returns a list of urls that the server is accessible from.
func (c Config) Addresses(sl *slog.Logger) error {
	const format = "config addresses: %w"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf(format, err)
	}

	if err := c.addresses(sl); err != nil {
		return fmt.Errorf(format, err)
	}
	return nil
}

// addresses prints a list of URLs that the server is accessible from.
func (c Config) addresses(sl *slog.Logger) error {
	const intro = "Depending on the firewall and operating system setup, " +
		"the Defacto2 web server application maybe accessible from these links:"
	sl.Info(intro)

	// Access fields directly without reflection
	port := uint64(c.HTTPPort)
	tls := uint64(c.TLSPort)

	if port == 0 && tls == 0 {
		return ErrNoPort
	}

	hosts, err := helper.LocalHosts()
	if err != nil {
		return fmt.Errorf("the server cannot get the local host names: %w", err)
	}

	const (
		disable = 0
		port80  = 80
		port443 = 443
		msg     = "URL"
	)

	matchHostStr := c.MatchHost.String()

	for host := range slices.Values(hosts) {
		if matchHostStr != "" && host != matchHostStr {
			continue
		}

		// Process HTTP
		if port != disable {
			var s string
			if port == port80 {
				s = "http://" + host
			} else {
				s = "http://" + net.JoinHostPort(host, strconv.FormatUint(port, 10))
			}
			sl.Info(msg, slog.Uint64("port", port), slog.String("link", s))
		}

		// Process HTTPS
		if tls != disable {
			var s string
			if tls == port443 {
				s = "https://" + host
			} else {
				s = "https://" + net.JoinHostPort(host, strconv.FormatUint(tls, 10))
			}
			sl.Info(msg, slog.Uint64("port", tls), slog.String("link", s))
		}
	}

	if matchHostStr != "" {
		return nil
	}

	ips, err := helper.LocalIPs()
	if err != nil {
		return fmt.Errorf("the server cannot get the local IP addresses: %w", err)
	}

	if port != disable {
		for ip := range slices.Values(ips) {
			s := "http://" + net.JoinHostPort(ip.String(), strconv.FormatUint(port, 10))
			sl.Info(msg, slog.Uint64("port", port), slog.String("link", s))
		}
	}

	return nil
}

// googleIDs uses vague formatting to mask the configuration so it is not printed or logged.
func (c Config) googleIDs(sl *slog.Logger, key string) {
	if sl == nil {
		return
	}

	accounts := c.GoogleAccounts
	helper, ok := any(accounts).(Helper)
	if !ok {
		return
	}
	tip := helper.Help()
	if tip == "" {
		return
	}

	const name = "GoogleAccounts"
	msg := Configured(name, accounts.String()) + ":"
	sl.Info(msg, slog.Any(key, mask), slog.String("tip", tip))
}

// StaticThumb returns the path to the thumbnail directory.
func StaticThumb() string {
	return "/public/image/thumb"
}

// StaticOriginal returns the path to the image directory.
func StaticOriginal() string {
	return "/public/image/original"
}
