// Package postgres connects to and interacts with a PostgreSQL database server.
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

	"github.com/caarlos0/env/v10"
	_ "github.com/jackc/pgx/v5/stdlib" // Use a lowlevel PostgreSQL driver.
	"go.uber.org/zap"
)

var (
	ErrEnv = errors.New("environment variable probably contains an invalid value")
	ErrZap = errors.New("zap logger instance is nil")
)

const (
	EnvPrefix  = "PS_"                  // EnvPrefix is the prefix for all server environment variables.
	DockerHost = "host.docker.internal" // DockerHost is the hostname of the internal Docker container.
	DriverName = "pgx"                  // DriverName of the database.
	Protocol   = "postgres"             // Protocol of the database driver.
)

// Connection details of the PostgreSQL database connection.
type Connection struct {
	HostName string `env:"HOST_NAME" envDefault:"localhost" help:"Host name of the database server"`
	Database string `env:"DATABASE" envDefault:"defacto2-ps" help:"The name of the database to connect to"`
	Username string `env:"USERNAME" help:"Database username used to connect"`
	Password string `env:"PASSWORD" help:"Password for the database username"`

	Protocol  string // Protocol scheme of the PostgreSQL database. Defaults to postgres.
	HostPort  int    `env:"HOST_PORT" envDefault:"5432" help:"Port number the Postgres database server is listening on"`
	NoSSLMode bool   `env:"NO_SSL" envDefault:"true" help:"Connect to the database using an insecure, plain text connection"`
}

// Open opens a PostgreSQL database connection.
func (c Connection) Open() (*sql.DB, error) {
	conn, err := sql.Open(DriverName, c.URL())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Check the connection values and print any issues or feedback to the logger.
func (c Connection) Check(logger *zap.SugaredLogger, local bool) error {
	if logger == nil {
		return ErrZap
	}
	if c.HostName == "" {
		logger.Warn("The database connection host name is empty.")
	}
	if c.HostPort == 0 {
		logger.Warn("The database connection host port is set to 0.")
	}
	if !local && c.NoSSLMode {
		logger.Warn("The database connection is using an insecure, plain text connection.")
	}
	switch {
	case c.Username == "" && c.Password != "":
		logger.Info("The database connection username is empty but the password is set.")
	case c.Username == "":
		logger.Info("The database connection username is empty.")
	case c.Password == "":
		logger.Info("The database connection password is empty.")
	}
	return nil
}

// New initializes the connection with default values or values from the environment.
func New() (Connection, error) {
	c := Connection{}
	c.Protocol = Protocol
	if err := env.ParseWithOptions(
		&c, env.Options{Prefix: EnvPrefix}); err != nil {
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
//
// An example connection "postgres://username:password@localhost:5432/postgres?sslmode=disable"
func (c Connection) URL() string {
	if c.Protocol == "" {
		c.Protocol = Protocol
	}
	var usr *url.Userinfo
	if c.Username != "" && c.Password != "" {
		usr = url.UserPassword(c.Username, c.Password)
	} else if c.Username != "" {
		usr = url.User(c.Username)
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
		h3       = "Environment variable"
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
		help := field.Tag.Get("help")
		if help == "" {
			continue
		}
		val := values.FieldByName(field.Name)
		id := field.Name
		name := EnvPrefix + field.Tag.Get("env")
		lead := func() {
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, val, help)
		}
		if id == "Password" && val.String() == c.Password {
			fmt.Fprintf(w, "\t%s\t%s\t%v\t%s.\n", id, name, "******", help)
			continue
		}
		lead()
	}
	nl()
	w.Flush()
	return b
}
