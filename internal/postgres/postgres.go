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
	DefaultURL = "postgres://root:example@localhost:5432/defacto2_ps" //nolint: gosec
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
		return Connection{}, fmt.Errorf("default url %w: %w", ErrEnv, err)
	}
	return c, nil
}

func Connections() (int64, int64, error) {
	// SELECT * FROM pg_stat_activity where datname='defacto2_ps';
	conn, err := ConnectDB()
	if err != nil {
		return 0, 0, fmt.Errorf("postgres connect, %w", err)
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT 'dataname' FROM pg_stat_activity WHERE datname='defacto2_ps';")
	if err != nil {
		return 0, 0, fmt.Errorf("postgres query, %w", err)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("postgres rows, %w", err)
	}
	defer rows.Close()
	count := int64(0)
	for rows.Next() {
		count++
	}
	max, err := conn.Query("SHOW max_connections;")
	if err != nil {
		return 0, 0, fmt.Errorf("postgres query, %w", err)
	}
	if err := max.Err(); err != nil {
		return 0, 0, fmt.Errorf("postgres rows, %w", err)
	}
	defer max.Close()
	var maxConnections int64
	for max.Next() {
		if err := max.Scan(&maxConnections); err != nil {
			return 0, 0, fmt.Errorf("postgres scan, %w", err)
		}
	}
	return count, maxConnections, nil

}

// ConnectDB connects to the PostgreSQL database.
// The connection must be closed after use.
func ConnectDB() (*sql.DB, error) {
	dataSource, err := New()
	if err != nil {
		return nil, fmt.Errorf("postgres new connection, %w", err)
	}
	conn, err := sql.Open(DriverName, dataSource.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres open new connection, %w", err)
	}
	return conn, nil
}

// ConnectTx connects to the PostgreSQL database and starts a transaction.
// The transaction must be committed or rolled back and the connection closed.
func ConnectTx() (*sql.DB, *sql.Tx, error) {
	conn, err := ConnectDB()
	if err != nil {
		return nil, nil, fmt.Errorf("postgres connect transaction, %w", err)
	}
	tx, err := conn.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("postgres begin transaction, %w", err)
	}
	return conn, tx, nil
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
