// Package postgres connects to and interacts with the PostgreSQL database server.
// The functions are specific to the Postgres platform rather than more generic or
// interchangeable SQL statements.
// The postgres/models directory is generated by SQLBoiler and should not be modified.
package postgres

import (
	"database/sql"
	"fmt"
)

const (
	Name = "postgres"
)

// ConnectDB connects to the PostgreSQL database.
func ConnectDB() (*sql.DB, error) {
	const (
		protocol = "postgres"
		user     = "root"
		password = "example"
		hostname = "localhost"
		hostport = 5432
		database = "defacto2-ps"
		options  = "sslmode=disable"
	)
	dsn := fmt.Sprintf("%s://%s:%s@%s:%d/%s?%s",
		protocol, user, password, hostname, hostport, database, options)
	conn, err := sql.Open(Name, dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Version returns the PostgreSQL database version from an SQL query.
func Version() (string, error) {
	conn, err := ConnectDB()
	if err != nil {
		return "", err
	}
	rows, err := conn.Query("SELECT version();")
	if err != nil {
		return "", err
	}
	var s string
	for rows.Next() {
		rows.Scan(&s)
	}
	rows.Close()
	conn.Close()
	return s, nil
}