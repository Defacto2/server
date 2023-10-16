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
	- Complete "Search for files" feature to support, Years, Descriptions.
	- Uploader placeholder.
	- Complete the textfile printing feature.
	- Fix missing warnings for the non-server commands of "address" and "config".
	- View file, with a magazine title, handle solo issue numbers.
	- About file screenshot should be stretched to match the width of the thumbnails.
	- Start repair should delete mimetype that begins with "ERROR: " (/f/a92a225).
	- Mime type  	Zip archive data, at least v1.0 to extract and v2.0 should be simplified.
	- Cache Pouet reviews and link to Demozoo ID if found.
	- If screenshot cannot fix side tab, CSS move them to the bottom of the about.
	- - Or create a modal popup for the screenshots that centre on the screen.
	- NFO preview isn't rendering high ASCII characters correctly.

# Mobile fixes

	- none

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
	- Move the glossary of terms from a module to its own page.

# Bugs

	- none

# New features to deliver

	- Extensive database repair and cleanup on startup.
	- DOS emulation using something other than DOSBox.

# Milestones to add

	- Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP
*/
// [Defacto2]: https://defacto2.net
// [Defacto2 PostgreSQL database]: https://github.com/Defacto2/database-ps
package main
