package postgres

import (
	"database/sql"
)

const (
	Name = "postgres"
)

func ConnectDB() (*sql.DB, error) {
	conn, err := sql.Open(Name, "postgres://root:example@localhost:5432/defacto2-ps?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

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
