// Copyright © 2023 Ben Garrett. All rights reserved.

/*
The [Defacto2] web server created in 2023 on Go.

Usage:

	df2-server

Launch the server and listen on the configured port (default: 1323).
The server expects the [Defacto2 PostgreSQL database] running on the host system
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

This application expects the [Defacto2 PostgreSQL database] and if needed,
the following environment variables to be set:

	1. "DEFACTO2_PORT" is the unencrypted port number the web server will listen on (default: 1323).
	2. "DEFACTO2_PORTS" is the encrypted port number the web server will listen on (default: 0 [unused]).
	3. "DEFACTO2_TIMEOUT" is the he timeout in seconds for the web server and database requests (default: 5).

# File downloads

The following environment variables are required for the web server to
offer file downloads:

	1. "DEFACTO2_DOWNLOAD" is the absolute path to the file downloads directory.

# Previews and thumbnails

The following environment variables are required for the previews and thumbnails:

	1. "DEFACTO2_SCREENSHOTS" is the absolute path to the screenshots directory.
	2. "DEFACTO2_THUMBNAILS" is the absolute path to the thumbnails directory.

# Dependencies

The following on the host system or in the container.

Coming soon.

# Configuration overrides

A number of server configuration options can be overridden using hard coded values.
Though these are not advised other than for debugging or testing in development.

The following options can be added to [Override].

	configs.IsProduction = true		// This will enable the production logger
	configs.HTTPSRedirect = true	// This requires HTTPS certificates to be installed and configured
	configs.NoRobots = true			// This will disable search engine crawling
	configs.LogRequests = true		// This will log all HTTP requests to the server or stdout

# Tasks

	- Finish this doc.go file.
	- (long) group/releaser pages should have a link to the end of the document.


# Mobile fixes

	- none

# Bugs

	-

# New features to deliver

	-

# Database changes

	- [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".

# Milestones to add or fix

	- Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP
*/
// [Defacto2]: https://defacto2.net
// [Defacto2 PostgreSQL database]: https://github.com/Defacto2/database-ps
package main
