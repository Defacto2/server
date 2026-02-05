# Defacto2, <small>website software</small>

[![Go Reference](server.svg)](https://pkg.go.dev/github.com/Defacto2/server)
[![License](license.svg)](../LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Defacto2/server)](https://goreportcard.com/report/github.com/Defacto2/server)

```
      ·      ▒██▀ ▀       ▒██▀ ▀              ▀ ▀▒██             ▀ ▀███ ·
      : ▒██▀ ▓██ ▒██▀▀██▓ ▓██▀▀▒██▀▀███ ▒██▀▀██▓ ▓██▀ ▒██▀▀███ ▒██▀▀▀▀▀ :
 · ··─┼─▓██──███─▓██─▄███─███──▓██──███─▓██──────███──▓██──███─▓██──███─┼─·· ·
      │ ███▄▄██▓ ███▄▄▄▄▄▄██▓  ███▄ ███▄███▄▄███ ███▄▄███▄ ███▄███▄▄███ │
· ··──┼─────────··                defacto2.net               ··─────────┼──·· ·
      │                                                                 :
```

The Defacto2 website is a self-contained application first devised in 2023.
It is built with the Go language and can be easily compiled for many platforms and operating systems.

All configurations are optional and any changes to the defaults are made through system environment variables.

## About

Defacto2 is a digital archive of online history and artifacts from the Scene—a once global community of programmers, artists, and others who create releases. The platform preserves and showcases:

- **Demos and intros** - Real-time computer graphics and music productions from the 1980s onward
- **Software artifacts** - Utilities, and applications from various computer platforms
- **Artist profiles** - Pages for individual sceners and groups (releasers) with their complete releases
- **File metadata** - Searchable database of thousands of artifacts with credits, descriptions, and external links
- **Historical preservation** - Links to related databases ([Demozoo](https://demozoo.org/), [Pouet](https://www.pouet.net/), [16colors](https://www.16colors.net/), expired websites, etc.)
- **Emulation support** - Play period-accurate DOS intros and demos via DOSBox emulation
- **Media previews** - View screenshots and download original files (when configured)

> [!NOTE]
> The application uses an _optional_ [PostgreSQL](https://www.postgresql.org/) database connection for data queries.
> While optional, you'll [need this database](https://github.com/Defacto2/database) running and [configured](docs/database.md) if you wish to browse the artifacts, releasers, and sceners.

## Architecture

Defacto2 is built as a full-stack application:

- **Backend**: Go web server using the [Echo](https://echo.labstack.com/) framework
- **Database**: [PostgreSQL](https://www.postgresql.org/) with type-safe queries via [SQLBoiler](https://github.com/aarondl/sqlboiler) ORM
- **Frontend**: Server-rendered HTML templates with [HTMX](https://htmx.org/) for interactivity
- **Assets**: Node.js-based CSS/JS compilation and minification pipeline
- **Configuration**: Entirely environment variable-based—no config files needed

The server is a single self-contained binary that embeds all static files, templates, and assets. It can optionally serve file downloads, image previews, and software emulation (DOS intros via DOSBox).

See the [Location Guide](docs/location.md) for the project structure and the [Source Setup](docs/source.md) guide for detailed architecture information.

## Download

[There are downloads available](https://github.com/Defacto2/server/releases/latest) for Linux, macOS and Windows.

## Installation

No installation is required to play around with the web server.

> [!NOTE]
> The Defacto2.net service does not currently share artifact files, screenshots, and thumbnails due to hosting costs.
> When self-hosted with asset directories configured (via `D2_DIR_DOWNLOAD`, `D2_DIR_PREVIEW`, `D2_DIR_THUMBNAIL` environment variables), these features are fully functional.

## Installation for the Debian Package

```sh
# download the latest release
$ wget https://github.com/Defacto2/server/releases/latest/download/defacto2-server_linux.deb

# install (or update) the package
$ sudo dpkg -i defacto2-server_linux.deb

# confirm the binary is executable
$ defacto2-server --version
```

## Usage

The web server will run without any arguments and will be available on the _[localhost](http://localhost:1323)_ over port `1323`.

```sh
$ defacto2-server

> ⇨ http server started on [::]:1323
```

To stop the server, press `CTRL+C`.

```
> Detected Ctrl + C, server will shutdown now
```

## Development Setup

To build and develop the Defacto2 server locally, you'll need:

- **[Go](https://go.dev/doc/install)** 1.25.5 or later
- **[Task](https://taskfile.dev/installation/)** - Task runner for building and testing
- **[Node.js](https://nodejs.org/)** / pnpm - For managing frontend dependencies

### Quick start

```sh
# Clone the repository
$ git clone https://github.com/Defacto2/server.git
$ cd server

# Install development dependencies (one-time setup)
$ task _init

# Run the development server with live reload
$ task serve
```

### Common development commands

```sh
task test       # Run all tests
task testr      # Run tests with race detection
task lint       # Format and lint all code
task binary     # Build a standalone binary
```

For more detailed development instructions, see the [Source Setup](source.md) guide.

## Configuration

The application uses environment variables to configure the database connection and other settings. These are documented in the [software package documentation](https://pkg.go.dev/github.com/Defacto2/server). 

There are examples of the environment variables in the [example .env](../init/example.env.local) and the [example .service](../init/defacto2.service) files found in the `init/` directory.

## Documentation

For developers and contributors:

- **[Source Setup](source.md)** - How to set up and build the project locally
- **[Database Guide](database.md)** - PostgreSQL setup, troubleshooting, and schema information
- **[Code Patterns](patterns.md)** - Go language patterns, SQLBoiler ORM examples, and development conventions
- **[Location Guide](location.md)** - Project structure and file organization

