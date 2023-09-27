// Copyright © 2023 Ben Garrett. All rights reserved.

/*
The [Defacto2] webserver created in 2023 on Go.

Usage:

	df2-server [flags]

The flags are:

	--help
			Print help and exit.
	--version
			Print version and exit.

# Database

This application requires a PostgreSQL database and the following environment
variables to be set:

	1. "DEFACTO2_DB_USER"
	2. "DEFACTO2_DB_PASS"
	3. "DEFACTO2_DB_NAME"
	4. "DEFACTO2_DB_HOST"
	5. "DEFACTO2_DB_PORT"

The following environment variables are optional:

	1. "DEFACTO2_DB_SSLMODE"
	2. "DEFACTO2_DB_SSLROOTCERT"
	3. "DEFACTO2_DB_SSLCERT"
	4. "DEFACTO2_DB_SSLKEY"

# Dependencies

The following on the host system or in the container.

The following environment variables are required for the Pouet API:

*/
// [Defacto2]: https://defacto2.net
package main
