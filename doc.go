// Copyright © 2023-2024 Ben Garrett. All rights reserved.

/*
The [Defacto2] web server is a self-contained application, first created in 2023 and built with the Go language.
And can be easily compiled for major operating systems.

The web server relies on a [PostgreSQL database] for data queries, best provided using a container such as [Docker].

All configurations and settings for the web application are through system environment variables.
Variables are handled within the container's environment on a production setup, such as with a Docker container.

# Installation

	*add installation instructions here*

# Usage

Usage:

	df2-server

Launch the server and listen on the configured port (default: 1323).
The server expects the Defacto2 [PostgreSQL database] running on the host system
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

This application expects the Defacto2 [PostgreSQL database] and the following environment variables to be set if needed:

	1. PS_USERNAME is the PostgreSQL account username.
	2. PS_PASSWORD is the PostgreSQL account password.

The following variables are optional:

	1. PS_HOST_NAME is the PostgreSQL server hostname (default: localhost).
	2. PS_HOST_PORT is the PostgreSQL server port number (default: 5432).
	3. PS_DATABASE is the PostgreSQL database name (default: defacto2-ps).
	4. PS_NO_SSL is the PostgreSQL connection is insecure and in plaintext (default: true).

# File serving

The following environment variables are used for the webserver to offer file downloads, software emulation, display previews and thumbnails:

	1. D2_DOWNLOAD is the absolute path to the file downloads directory.
	2. D2_SCREENSHOTS is the absolute path to the screenshots directory.
	3. D2_THUMBNAILS is the absolute path to the thumbnails directory.

# Web server

Finally, a couple of environment variables change the server-specific options.

	1. D2_HTTP_PORT is the unencrypted port number the web server will listen on (default: 1323).
	2. D2_LOG_REQUESTS is the web server will log all HTTP requests to stdout (default: false).

# Using the source code

The repository configurations use [Task] for binary compiling, which needs local installation.

A new cloned repository needs to download dependencies.

	task _init

The list of available tasks can be shown.

	task --list-all (or just task)

To run a local server with live reloading, reflecting any source code changes.
The task uses the `.env.local` file for configurations which should be in the repository root directory.
A `example.env.local` file is provided as a template.

	task serve

To reflect any changes to the JS or CSS files, a task is available to minify and copy the assets.

	task assets

# Building the source code

To build a binary for the local machine.

	task build

	# run the binary
	./df2-server --version

To build a collection of binaries for various platforms.
The resulting packages are in the dist directory in the repository root.

	build-release

	# or if the source code has changed
	build-snapshot

	# list the contents of the dist directory
	ls -l dist/

# Lint source code changes

The application is configured to use [golangci-lint] as the Go linter aggregator.

	task lint

*/
// [Defacto2]: https://defacto2.net
// [PostgreSQL database]: https://github.com/Defacto2/database-ps
// [Docker]: https://www.docker.com/products/docker-desktop
// [Task]: https://taskfile.dev/installation
// [golangci-lint]: https://golangci-lint.run/usage/install/#local-installation
package main
