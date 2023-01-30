# Defacto2 server

The Defacto2 web server is a self-contained application built in Go. It can be quickly and easily built for all the common operating systems. The web server relies on a PostgreSQL database server for data queries. This is best provided using a container such as Docker.

All configurations and settings for this web server are handled through system environment variables. On a production setup, this too should be hosted within a container such as Docker.

### TODOs

- [ ] All SQL smts need to account for `delete_at`
`qm.WithDeleted`
- [ ] Tests for routes and templates.

### Possible TODOs

- [ ] Create a method to calc the most popular years for a collection of records query.
- [ ] `OrderBy` Name/Count /html3/groups?
https://pkg.go.dev/sort#example-package-SortKeys
- [ ] Move `OrderBy` params to cookies?

---

## Developer

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

to fill out

---