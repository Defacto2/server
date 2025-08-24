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

The application uses an optional [PostgreSQL](https://www.postgresql.org/) [database connection](https://www.postgresql.org/docs/current/ecpg-sql-connect.html) for data queries.

All configurations are optional and any changes to the defaults are made through system environment variables.

## Download

[There are downloads available](https://github.com/Defacto2/server/releases/latest) for Linux, macOS and Windows.

## Installation

No installation is required to play around with the web server.

> [!NOTE]
> Currently the artifact screenshots, thumbnails and files are not available for download or display. 
> This due to the additional hosting costs.

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

The web server will run with out any arguments and will be available on the _[localhost](http://localhost:1323)_ over port `1323`.

```sh
$ defacto2-server

> ⇨ http server started on [::]:1323
```

To stop the server, press `CTRL+C`.

```
> Detected Ctrl + C, server will shutdown now
```


## Configuration

The application uses environment variables to configure the database connection and other settings. These are documented in the [software package documentation](https://pkg.go.dev/github.com/Defacto2/server). 

There are examples of the environment variables in the [example .env](../init/example.env.local) and the [example .service](../init/defacto2.service) files found in the `init/` directory.

## Source code

[How to use the source code](https://github.com/Defacto2/server/blob/main/docs/source.md).
