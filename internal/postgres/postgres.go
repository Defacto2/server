// Package postgres connects to and interacts with the PostgreSQL database server.
// The functions are specific to the Postgres platform rather than more generic or
// interchangeable SQL statements.
// The postgres/models directory is generated by SQLBoiler and should not be modified.
package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/caarlos0/env/v7"
	_ "github.com/jackc/pgx/v5/stdlib" // Use a lowlevel PostgreSQL driver.
)

var ErrEnv = errors.New("environment variable probably contains an invalid value")

const (
	Protocol   = "postgres"  // Protocol of the database driver.
	DriverName = "pgx"       // DriverName of the database.
	User       = "root"      // User is the default database username used to connect.
	Pass       = "example"   // Pass is the placeholder database password used to connect.
	HostName   = "localhost" // HostName is the default host name of the database server to connect.
	HostPort   = 5432        // HostPort is the default port number of the database server to connect.
	// DockerHost is the default database host name to use when running in a Docker container.
	DockerHost = "host.docker.internal"
	DBName     = "defacto2-ps" // DBName is the default database name to connect to.
	NoSSL      = true          // NoSSL connects to the database using an insecure, plain text connection.
)

// Connection details of the PostgreSQL database connection.
type Connection struct {
	// Protocol scheme of the PostgreSQL database. Defaults to postgres.
	Protocol string
	// HostName is the host name of the server. Defaults to localhost.
	HostName string `env:"PS_HOST" help:"Host name of the database server"`
	// HostPort is the port number the server is listening on. Defaults to 5432.
	HostPort int `env:"PS_PORT" help:"Port number the Postgres database server is listening on"`
	// Database is the database name.
	Database string `env:"PS_DB" help:"Database name to connect to"`
	// NoSSLMode connects to the database using an insecure,
	// plain text connection using the sslmode=disable param.
	NoSSLMode bool `env:"PS_NO_SSL" help:"Connect to the database using an insecure, plain text connection"`
	// User is the database user used to connect to the database.
	User string `env:"PS_USER" help:"Database user name used to connect"`
	// Password is the password for the database user.
	Password string `env:"PS_PASS" help:"Password for the database user"`
}

// Open opens a PostgreSQL database connection.
func (c Connection) Open() (*sql.DB, error) {
	conn, err := sql.Open(DriverName, c.URL())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// New initializes the connection with default values or values from the environment.
func New() (Connection, error) {
	// "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	c := Connection{}
	c.NoSSLMode = NoSSL
	c.Protocol = Protocol
	c.User = User
	c.Password = Pass
	c.HostName = DockerHost
	c.HostPort = HostPort
	c.Database = DBName
	if err := env.Parse(&c, env.Options{}); err != nil {
		return Connection{}, fmt.Errorf("%w: %w", ErrEnv, err)
	}
	return c, nil
}

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() (*sql.DB, error) {
	ds, err := New()
	if err != nil {
		return nil, err
	}
	conn, err := sql.Open(DriverName, ds.URL())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// URL returns a url used as a PostgreSQL database connection.
func (c Connection) URL() string {
	// example url string:
	// "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	if c.Protocol == "" {
		c.Protocol = Protocol
	}
	if c.HostName == "" {
		c.HostName = HostName
	}
	if c.HostPort < 1 {
		c.HostPort = HostPort
	}
	var usr *url.Userinfo
	if c.User != "" && c.Password != "" {
		usr = url.UserPassword(c.User, c.Password)
	} else if c.User != "" {
		usr = url.User(c.User)
	}
	dns := url.URL{
		Scheme: c.Protocol,
		User:   usr,
		Host:   fmt.Sprintf("%s:%d", c.HostName, c.HostPort),
		Path:   c.Database,
	}
	if c.NoSSLMode {
		q := dns.Query()
		q.Set("sslmode", "disable")
		dns.RawQuery = q.Encode()
	}
	return dns.String()
}

// Configurations prints a list of active connection configurations.
func (c Connection) Configurations(b *strings.Builder) *strings.Builder {
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

	fields := reflect.VisibleFields(reflect.TypeOf(c))
	values := reflect.ValueOf(c)
	w := tabwriter.NewWriter(b, minwidth, tabwidth, padding, padchar, flags)
	nl := func() {
		fmt.Fprintf(w, "\t\t\t\t\n")
	}

	fmt.Fprint(b, "PostgreSQL database connection configuration.\n\n")
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
		val := values.FieldByName(field.Name)
		id := field.Name
		name := field.Tag.Get("env")
		help := field.Tag.Get("help")
		if help == "" {
			continue
		}
		lead := func() {
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, val, help)
		}
		if id == "Password" && val.String() != Pass {
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, "****", help)
			continue
		}
		lead()
	}
	nl()
	w.Flush()
	return b
}
