package main

/*
	SQLBoiler command to generate Go code from a PostgreSQL database schema.
	https://github.com/volatiletech/sqlboiler

	It requires an active PostgreSQL server to be running.

	To rebuild run this command in the terminal:
	$ go generate

	--config ".sqlboiler.toml"		- Use the configuration file "sqlboiler.toml".
	--wipe							- Wipe any existing generated files before re-generation.
	--add-soft-deletes				- [REQUIRED] Add soft delete support to the generated models.
	psql							- Use the PostgreSQL database driver.
*/

//go:generate sqlboiler --config "init/.sqlboiler.toml" --wipe --add-soft-deletes psql
