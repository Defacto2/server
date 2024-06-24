// Copyright © 2023-2024 Ben Garrett. All rights reserved.

/*

The [Defacto2] application is a self-contained web server first devised in 2023.
It is built with the Go language and can be easily compiled for significant server operating systems.
The application relies on a [PostgreSQL] database setup for data queries using a PostgreSQL [database connection].

All configurations and modifications to this web application's default settings are through system environment variables.

While you can compile the application to target Windows environments, it is ill-advised as it needs to work correctly with NTFS file paths.
Instead, it is advisable to use Windows Subsystem for Linux.

# Usage

While it will not be fully functional without directory paths or a database connection, the web server will work out of the box without any configuration provided.

Usage:

	defacto2-server

The web server should be available at the unencrypted address, http://localhost:1323.

# Commands

There are only two additional commands: one lists the accessible addresses the web server listens on, and the other lists the detected settings.

Usage:

	defacto2-server [command]

The commands are:

	address
			List the IP, hostname and port addresses the server is listening on.
	config
			List the server configuration options and settings.

# Flags

Usage:

	defacto2-server [flags]

The flags are:

	--help
			Print the basic command help and exit.
	--version
			Print the application version and exit.

# Configuration

All application configurations and default modifications are made with environment variables.
In systemd, these can be provided using a system service (unit) file, using the Environment assignment under the Service type.

You can use an example of a defacto2.service unit file, which is found in the source code repository in the init/ directory.

A partial example of a defacto2.service unit file:

	[Unit]
	Description=Defacto2

	[Service]
	Environment="D2_MATCH_HOST=localhost"
	Environment="D2_DATABASE_URL=postgres://root:example@localhost:5432/defacto2_ps"
	Environment="D2_PROD_MODE=true" "D2_READ_ONLY=false" "D2_NO_CRAWL=true"

# Database

This application expects a PostgreSQL database with the Defacto2 "files" table and the connection URL configured in the D2_DATABASE_URL environment variable.

The [database connection] URL uses a configuration by default but will fail to connect unless the testing Postgres database matches the same values.
In production, you must provide a secure and working connection D2_DATABASE_URL variable.
The default database connection URL is: postgres://root:example@localhost:5432/defacto2_ps

Some examples:
	// local connection
	D2_DATABASE_URL=postgres://username:password@localhost:5432/database_name

	// Docker connection
	D2_DATABASE_URL=postgres://username:password@host.docker.internal:5432/database_name

# File assets

The web server uses the following environment variables to offer file downloads, software emulation, web server previews, and thumbnails.
All paths must be absolute and valid and must contain tens of thousands of asset files named with universal unique identifiers.
The website will turn off the associated feature if the provided path is invalid or contains unexpected files.

	- D2_DIR_DOWNLOAD is the absolute path to the file downloads directory.

	- D2_DIR_PREVIEW is the absolute path to the image screenshots directory.

	- D2_DIR_THUMBNAIL is the absolute path to the squared, image thumbnails directory.

	- D2_DIR_ORPHANED is the absolute path to the orphaned files directory.

	- D2_DIR_EXTRA is the absolute path to the extra files directory.

An example download setting:

	D2_DIR_DOWNLOAD=/mnt/volume/assets/downloads

# Log file storage

When the application runs in production mode, errors or warnings caused by the web server are saved to a log file.
The location can be provided using the D2_DIR_LOG variable, which must point to an absolute directory path.
Otherwise, the server will create a subdirectory using [os.UserConfigDir].

An example log setting:

	D2_DIR_LOG=/var/log/defacto2-server

# Administrator accounts

The web server uses [Google OAuth2] for administrator logins.
The server requires a Google OAuth2 client ID to validate admin logins, which is provided in the D2_GOOGLE_CLIENT_ID environment variable.

The server also requires a list of Google OAuth2 user accounts, which is provided in the D2_GOOGLE_ACCOUNTS environment variable.
A user account is the JWT ["sub"] field assertion in the form of a unique integer.

An example accounts setting:

	D2_GOOGLE_CLIENT_ID=123-abc.apps.googleusercontent.com
	D2_GOOGLE_ACCOUNTS=1234567890,0987654321

# Production mode

The production mode is on by default and should be enabled in production as it has the following effects.

	1. It runs file assets and database entry checks on startup.
	2. Any errors or warnings get appended to a log file.
	3. If the server crashes, it will recover instead of exiting the program.
	4. Turns off the Uploader form debug feature.
	5. Force the administrator logins to be served only over encrypted HTTPS protocols.

To turn off production mode:

	D2_PROD_MODE=false

# Read-only mode

Read-only mode blocks any website feature that writes to the server database, including the Uploader and administrator database entry edits.

To turn off the read-only mode:

	D2_READ_ONLY=true

# No crawl mode

The no crawler mode inserts an [X-Robots-Tag] with the "none" value for all network response headers sent by the web server.
The header advises search engines and other bots not to index the website or assets.

To enable no crawl mode:

	D2_NO_CRAWL=true

# Quiet startup

When quiet is turned on, the majority of startup messages are suppressed.
This option is meant for [systemd] to avoid spamming its log.

To enable quiet mode:

	D2_QUIET=true

# HTTP and HTTPS

The web server will listen to all HTTP requests on port 1323 without configuration.
The value can be changed with the D2_HTTP_PORT variable, which can be set to 0 to disable HTTP.

The D2_TLS_PORT variable, which is turned off by default, allows for an encrypted HTTPS service.
In a production situation, you should supply the D2_TLS_CERT and D2_TLS_KEY variables.
These should have an absolute path to a TLS (Transport Layer Security) certificate file and key file.
If no certificate or key is provided, a dummy certificate will be used, but browsers will reject these.

If the D2_HTTP_PORT and the D2_TLS_PORT values are set to 0, the web server will override to enable port 1323 for HTTP connections.
On Linux, ports 1-1023 are considered well-known and reserved for the operating system.

Providing a D2_MATCH_HOST variable can restrict the web server from listening to HTTP and HTTPS requests from a single IP address or host.
Otherwise, the web server listens to all requests on the ports.

An example configuration to exclusively use HTTPS and only accept local connections:

	D2_HTTP_PORT=0
	D2_MATCH_HOST=localhost
	D2_TLS_PORT=443
	D2_TLS_CERT=/etc/ssl/certs/localhost.crt
	D2_TLS_KEY=/etc/ssl/private/localhost.key

*/
// [Defacto2]: https://defacto2.net
// [PostgreSQL]: https://github.com/Defacto2/database-ps
// [database connection]: https://www.postgresql.org/docs/current/ecpg-sql-connect.html
// [Ubuntu Server]: https://ubuntu.com/server
// [Docker]: https://www.docker.com/products/docker-desktop
// [Task]: https://taskfile.dev/installation
// [golangci-lint]: https://golangci-lint.run/usage/install/#local-installation
// [Google OAuth2]: https://developers.google.com/identity/account-linking/oauth-with-sign-in-linking
// ["sub"]: https://developers.google.com/identity/account-linking/oauth-with-sign-in-linking#validate_and_decode_the_jwt_assertion
// [X-Robots-Tag]: https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag
// [systemd]: https://systemd.io/
package main
