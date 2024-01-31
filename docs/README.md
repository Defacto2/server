# Defacto2, web application server

The [Defacto2](https://defacto2.net) web server is a self-contained application built on [Go](https://go.dev/). 
It can be quickly and easily built for all the common operating systems. 
The web server relies on a [PostgreSQL database server](https://www.postgresql.org/) for data queries. 
This is best provided using a container such as [Docker](https://www.docker.com/).

All configurations and settings for the web server are handled through system environment variables. 
On a production setup such with a Docker container, these variables should also be handled within the container environment.

# TODOs

These items should be implemented and tested before going live.

#### Database

- [ ] Database *normalisations* on server start. 
- - [ ] PostgreSQL *indexes* with case-sensitive strings.
- [ ] Some form of database timeout? **The Echo middleware is problematic.**
- [ ] All SQL statements need to account for `delete_at` ~ `qm.WithDeleted`
---

#### Pages

- [ ] Show a *custom error page when the file download* is missing or the root directory is broken.
- [ ] Run an *automated test to confirm 200 status* for all routes. Run this on startup using a defer func?
- [ ] Tests for routes and templates.

#### Automatic database corrections

- `/g/damn-excellent-ansi-designers` > `damn-excellent-ansi-design`
- `/g/the-original-funny-guys` > `original-funny-guys`

#### Known broken links

- [ ] http://localhost:1323/g/x_pression-design
- [ ] http://localhost:1323/g/ice
- [ ] http://localhost:1323/g/ansi-creators-in-demand
- [ ] http://localhost:1323/g/nc_17
- [ ] http://localhost:1323/g/share-and-enjoy
- [ ] http://localhost:1323/g/north-american-pirate_phreak-association

### Possible TODOs

- [ ] `OrderBy` Name/Count /html3/groups?
https://pkg.go.dev/sort#example-package-SortKeys
- [ ] Move `OrderBy` params to cookies?
- [ ] (long) group/releaser pages should have a link to the end of the document.
- [ ] [model.Files.ListUpdates], rename the PSQL column from "updated_at" to "date_updated".
- [ ] Fetch the DOD nfo for w95, https://scenelist.org/nfo/DOD95C1H.ZIP

#### Support Unicode slug URLs as currently the regex removes all non alphanumeric chars.

```
/*
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/
```

# Install on Debian/Ubuntu

The following instructions uses the Debian packages management tool, 
`dpkg` to install the server software.

```sh
# Download the Debian package
wget https://github.com/Defacto2/server/releases/latest/download/df2-server_0.0.7_amd64.deb # todo need to rename

# Install (or update) the server
dpkg -i df2-server_0.0.7_amd64.deb

# Test the server
df2-server --version
df2-server --help

# Start the server without any custom configuration and in the developer mode
df2-server
```

# Edit the code

This web server is dependancy free and built in Go with some common frameworks and dependencies. 
The web application expects a local PostgreSQL server containing the Defacto2 database running on port `5432`.
The application will load and show static webpages, but any other page that requires data from database will timeout.

The web server configured to use the following as developer defaults.

- user: `root`
- password: `example`
- hostname: `localhost`
- database: `defacto2-ps`
- sslmode: `disabled`

[Download and install the Go](https://go.dev/doc/install) programming language.

Clone this repository using [git](https://git-scm.com/).

```sh
# Clone the repository
git clone https://github.com/Defacto2/server.git df2server
cd df2server

# Install some additional tools for use in the repository
# Task is a task runner / build tool
go install github.com/go-task/task/v3/cmd/task@latest

# Initalise the repository using the task runner
task _init

# Run the web application source code with a live code modification reload monitor.
task serve

...
⇨ Compiled with Go 1.21.0 for Windows on Intel/AMD 64.
⇨ http server started on [::]:1323
```

Point your browser to: **http://localhost:1323**.

To exit the server, tap <kbd>CTRL-C</kbd>.

```
00:00:00	DEBUG	df2-serve/server.go:58	The server is running in the development mode.

       ·      ▒██▀ ▀       ▒██▀ ▀              ▀ ▀▒██             ▀ ▀███ ·
       : ▒██▀ ▓██ ▒██▀▀██▓ ▓██▀▀▒██▀▀███ ▒██▀▀██▓ ▓██▀ ▒██▀▀███ ▒██▀▀▀▀▀ :
  · ··─┼─▓██──███─▓██─▄███─███──▓██──███─▓██──────███──▓██──███─▓██──███─┼─·· ·
       │ ███▄▄██▓ ███▄▄▄▄▄▄██▓  ███▄ ███▄███▄▄███ ███▄▄███▄ ███▄███▄▄███ │
 · ··──┼─────────··                defacto2.net               ··─────────┼──·· ·
       │                                                                 :

⇨ ...
⇨ http server started on [::]:1323

```

---

### Linter

The application is configured to use golangci-lint as the Go linter aggregator.

[Follow one of the local installation instructions](https://golangci-lint.run/usage/install/#local-installation).

```sh
# Use the lint task
task lint
```

---

### BootStrap 5

The site relies on Bootstrap for its frontend feature set. To avoid the messiness of JS package managers, any future Bootstrap updates can be manually sourced from the [Compiled CSS and JS Download](https://getbootstrap.com/docs/5.3/getting-started/download).

#### CSS:

`bootstrap.min.css` is located at: `/public/css/bootstrap.min.css`

#### JS:  

`bootstrap.bundle.min.js` is located at: `/public/js/bootstrap.bundle.min.js`

---

### GoReleaser

GoReleaser is an automation tool for Go projects.

https://goreleaser.com/

```sh
go install github.com/goreleaser/goreleaser@latest
```

The configuration file is found at `.goreleaser.yaml`.

To validate the file.

```sh
goreleaser check
```

To build a local-only release to confirm the builder configuration.

```sh
goreleaser release --snapshot --clean
```

To build to the local, host operating system. The compiled binary will be found in `dist/`.

```
goreleaser build --single-target
```

To run the build-only mode.

```
goreleaser build
```

Note, the `release` flag is unused, instead all new releases are compiled using GitHub Actions.
