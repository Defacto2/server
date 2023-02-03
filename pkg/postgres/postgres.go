// Package postgres connects to and interacts with the PostgreSQL database server.
// The functions are specific to the Postgres platform rather than more generic or
// interchangeable SQL statements.
// The postgres/models directory is generated by SQLBoiler and should not be modified.
package postgres

import (
	"database/sql"
	"fmt"
	"net/url"
)

const (
	// Name of the database driver.
	Name = "postgres"
)

// Connection details of the PostgreSQL database connection.
type Connection struct {
	Protocol string // Protocol scheme of the PostgreSQL database. Defaults to postgres.
	User     string // User is the database user used to connect to the database.
	Password string // Password is the password for the database user.
	HostName string // HostName is the host name of the server. Defaults to localhost.
	HostPort int    // HostPort is the port number the server is listening on. Defaults to 5432.
	Database string // Database is the database name.
	// NoSSLMode connects to the database using an insecure,
	// plain text connecction using the sslmode=disable param.
	NoSSLMode bool
}

// Open opens a PostgreSQL database connection.
func (c Connection) Open() (*sql.DB, error) {
	conn, err := sql.Open(Name, c.URL())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// URL returns a url used as a PostgreSQL database connection.
func (c Connection) URL() string {
	// "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	if c.Protocol == "" {
		c.Protocol = Name
	}
	if c.HostName == "" {
		c.HostName = "localhost"
	}
	if c.HostPort < 1 {
		c.HostPort = 5432
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

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() (*sql.DB, error) {
	dsn := Connection{
		User:      "root",
		Password:  "example",
		Database:  "defacto2-ps",
		NoSSLMode: true,
	}
	conn, err := sql.Open(Name, dsn.URL())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
