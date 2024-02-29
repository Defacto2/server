// SQLBoiler command to generate Go code from a PostgreSQL database schema.
// https://github.com/volatiletech/sqlboiler
//
// It requires an active PostgreSQL server to be running.
// To rebuild run this command in the terminal:
//
// $ go generate
package main

//go:generate sqlboiler --config "sqlboiler.toml" --wipe --add-soft-deletes psql
