// Package postgres connects to and interacts with a PostgreSQL database server.
// The functions are specific to the Postgres platform rather than more generic or
// interchangeable SQL statements.
//
// The postgres/models directory is generated by SQLBoiler and should not be modified.
package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v10"
	_ "github.com/jackc/pgx/v5/stdlib" // Use a lowlevel PostgreSQL driver.
	"go.uber.org/zap"
)

var (
	ErrEnv = errors.New("environment variable probably contains an invalid value")
	ErrZap = errors.New("zap logger instance is nil")
)

const (
	// DefaultURL is an example PostgreSQL connection string, it must not be used in production.
	DefaultURL = "postgres://root:example@localhost:5432/defacto2_ps"
	// DriverName of the database.
	DriverName = "pgx"
	// Protocol of the database driver.
	Protocol = "postgres"
)

// New initializes the connection with default values or values from the environment.
func New() (Connection, error) {
	c := Connection{
		URL: DefaultURL,
	}
	if err := env.Parse(&c); err != nil {
		return Connection{}, fmt.Errorf("%w: %w", ErrEnv, err)
	}
	return c, nil
}

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() (*sql.DB, error) {
	dataSource, err := New()
	if err != nil {
		return nil, fmt.Errorf("new connection db: %w", err)
	}
	conn, err := sql.Open(DriverName, dataSource.URL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	return conn, nil
}

// Connection details of the PostgreSQL database connection.
type Connection struct {
	URL string `env:"D2_DATABASE_URL"` // unsetting this value will cause the default to be used after a single use
}

// Validate the connection URL and print any issues to the logger.
func (c Connection) Validate(logger *zap.SugaredLogger) error {
	if logger == nil {
		return ErrZap
	}
	if c.URL == "" {
		logger.Warn("The database connection host name is empty")
	}
	u, err := url.Parse(c.URL)
	if err != nil {
		logger.Warn("The database connection URL is invalid, ", err)
	}
	if u.Scheme != Protocol {
		logger.Warnf("The database connection scheme is not: %s", Protocol)
	}
	return nil
}
