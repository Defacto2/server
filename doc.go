// Copyright © 2023 Ben Garrett. All rights reserved.

/*
The [Defacto2] web server created in 2023 on Go.

(requirements and dependencies should be mentioned here)

	- database
	- filedownloads
	- previews and thumbs
	- container or host system dependencies

Usage:

	df2-server

Launch the server and listen on the configured port (default: 1323).
The server expects the Defacto2 PostgreSQL database running on the host system
or in a container. But will run without a database connection for debugging.

Usage commands:

	df2-server [command]

The commands are:

	address
			List the IP, hostname and port addresses the server is listening on.
	config
			List the server configuration options and settings.

Usage flags:

	df2-server [flags]

The flags are:

	--help
			Print the basic command help and exit.
	--version
			Print the application version and exit.

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

# Configuration overrides

A number of server configuration options can be overridden by code edits.
Though these are not advised other than for debugging or testing in development.

The following options can be added to [github.com/Defacto2.server.main]

	configs.IsProduction = true		// This will enable the production logger

	configs.HTTPSRedirect = true	// This requires HTTPS certificates to be installed and configured

	configs.NoRobots = true			// This will disable search engine crawling

	configs.LogRequests = true		// This will log all HTTP requests to the server or stdout

# Tasks

	- Finish this doc.go file.
	- Hide Tables for a future implementation.
	- Fix "122 bbss (760k)"... "/bbs".
	- Trim Websites menu to only show List the Sites and Categories.
	- Hide the Mirror websites link.
	- Complete "Search for files" feature to support, Years, Descriptions.
	- Uploader placeholder.
	- Complete the textfile printing feature.
	- Database from MySQL to PostgreSQL migration and writeup.
	- Fix missing warnings for the non-server commands of "address" and "config".

# TODO

	- [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".
	- [model.RepairReleasers], globalize the "fixes" map and create redirects for the old names?
	- [handler.html3.Routes], using a range over with "echo.GET" does not work.
	- [handler.app], create a func for the aboutReadme.
	If it is a platform "amigatext" use topaz pre, else use filedownload, except for known archives.
	Also do a scan to confirm is not a binary file.
	- [handler.Configuration.Controller], handle a broken DB connection situation.
	- "conf.Import.ThumbnailDir" (dir path?) should be renamed to "/image/thumb".
	- [handler.Configuration.Moved] Implement legacy URI redirects,
	"/cracktros-detail.cfm:/:id" and "/code".

# Bugs

# New features to deliver

	- Extensive database repair and cleanup on startup.
	- DOS emulation using something other than DOSBox.

# Milestones to add

	- Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP
*/
// [Defacto2]: https://defacto2.net
package main
