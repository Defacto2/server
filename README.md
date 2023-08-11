# Defacto2 server

The Defacto2 web server is a self-contained application built in Go. It can be quickly and easily built for all the common operating systems. The web server relies on a PostgreSQL database server for data queries. This is best provided using a container such as Docker.

All configurations and settings for this web server are handled through system environment variables. On a production setup, this too should be hosted within a container such as Docker.

### TODOs

- [ ] Database normalisations on server start. PostgreSQL indexes with case-sensitive strings.
- [ ] TODO pass config.Timeout to the SQL contexts. There's no point echo timing out and the database query continuing.

```

[DIR] Acid Productions                           720    38M
[DIR] ACiD Productions                            10   515k

[DIR] Aces of ANSI Art                            85   809k
[DIR] Aces Of ANSI Art                             1   383k

[DIR] Addiction in Releasing                       3    18k
[DIR] Addiction In Releasing                       3    19k

```

- [ ] App downloads of files and a custom page when downloads directory is broken
- [ ] 
- [ ] Run an automated test to confirm 200 status for all routes. Run this on startup using a defer func?
- [ ] All SQL stmts need to account for `delete_at`
`qm.WithDeleted`
- [ ] Tests for routes and templates.
- [ ] Move contexts to the start of args.

### Automatic database corrections:

/g/damn-excellent-ansi-designers    > /g/damn-excellent-ansi-design
/g/the-original-funny-guys          > original-funny-guys


Broken links
- http://localhost:1323/g/x_pression-design
- http://localhost:1323/g/ice
- http://localhost:1323/g/ansi-creators-in-demand
- http://localhost:1323/g/nc_17
- http://localhost:1323/g/share-and-enjoy
- north-american-pirate_phreak-association

### Possible TODOs

- [ ] Create a method to calc the most popular years for a collection of records query.
- [ ] `OrderBy` Name/Count /html3/groups?
https://pkg.go.dev/sort#example-package-SortKeys
- [ ] Move `OrderBy` params to cookies?

Support Unicode slug URLs as currently the regex removes all non alphanumeric chars.

```
/*
Error:      	Not equal:
            	expected: "Mooñpeople"
            	actual  : "Moopeople"

				use utf8 lib to detect extended chars?
*/
```

---

## Install on Debian/Ubuntu

The following instructions uses the Debian packages management tool, `dpkg` to install the server software.

```sh
# Download the Debian package
wget https://github.com/Defacto2/server/releases/latest/download/df2-server_0.0.7_amd64.deb # todo need to rename

# Install (or update) the server
dpkg -i df2-server_0.0.7_amd64.deb

# Test the server
df2-server --version
df2-server --help

# Start the server in the developer mode
df2-server
```

---

## Edit the code

This web server is dependancy free and built in Go. 
The server expects a local PostgreSQL server containing the Defacto2 database running on port `5432`.
It is configured to use the following as developer defaults.

- user: `root`
- password: `example`
- hostname: `localhost`
- database: `defacto2-ps`
- sslmode: `disabled`

[Download and install Go](https://go.dev/doc/install).

Clone this repository using [git](https://git-scm.com/).

```sh
git clone https://github.com/Defacto2/server.git df2server
cd df2server
```

Compile and run the server.

```sh
go run .
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

⇨ Defacto2 web application with PostgreSQL 15.1.
⇨ 5 active routines sharing 4 usable threads on 4 CPU cores.
⇨ Compiled with Go 1.19.5.
⇨ http server started on [::]:1323

```

---

### Live reloading

The server is configured to use, [Air - Live reload for Go apps](https://github.com/cosmtrek/air).

Install Air to the server directory.

```
cd df2server
go install github.com/cosmtrek/air@latest

air
```

```
  __    _   ___ 
 / /\  | | | |_) 
/_/--\ |_| |_| \_ , built with Go 

mkdir /home/ben/github/df2-serve/tmp
watching .
building...
running...

server.go has changed
building...
running...
```

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
goreleaser release --snapshot --rm-dist
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

---

### Linter

[Follow one of the local installation instructions](https://golangci-lint.run/usage/install/#local-installation).

```sh
cd df2server
golangci-lint run ./...
```

---

### Gofumpt

```
cd df2server
go install mvdan.cc/gofumpt@latest
gofumpt -l -w .
```

---

### GCI

```
cd df2server
go install github.com/daixiang0/gci@latest
gci write ./..
```

---

### BootStrap 5

The site relies on Bootstrap for its frontend feature set. To avoid the messiness of JS package managers, any future Bootstrap updates can be manually sourced from the [Compiled CSS and JS Download](https://getbootstrap.com/docs/5.3/getting-started/download).


CSS: `bootstrap.min.css` is located at: `/public/css/bootstrap.min.css`

JS:  `bootstrap.bundle.min.js` is located at: `/public/js/bootstrap.bundle.min.js`