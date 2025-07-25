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
It is built with the Go language and can be easily compiled for significant server operating systems.
The application relies on a [PostgreSQL](https://www.postgresql.org/) database setup for data queries using a PostgreSQL [database connection](https://www.postgresql.org/docs/current/ecpg-sql-connect.html).

All configurations and modifications to this web application's default settings are through system environment variables.

## Download

Currently the application is available as a [standalone binary for Linux](https://github.com/Defacto2/server/releases/latest/download/defacto2-server_linux.zip).

## Installation

Installation instructions are provided for [Ubuntu Server](https://ubuntu.com/server). 

```sh
# download the latest release
wget https://github.com/Defacto2/server/releases/latest/download/defacto2-server_linux.deb

# install (or update) the package
sudo dpkg -i defacto2-server_linux.deb

# confirm the binary is executable
defacto2-server --version
```

For other Linux distributions, the binary can be installed manually to a directory in the system's PATH.

```sh
# download the latest release
wget https://github.com/Defacto2/server/releases/latest/download/defacto2-server_linux.zip

# unzip the archive
unzip defacto2-server_linux.zip

# make the binary executable
sudo chmod +x defacto2-server

# move the binary to a system path
sudo mv defacto2-server /usr/local/bin

# confirm the binary is executable
defacto2-server --version
```

## Usage

The web server will run with out any arguments and will be available on the _[localhost](http://localhost)_ with port `1323`.

```sh
defacto2-server
```

To stop the server, press `CTRL+C`.

## Configuration

The application uses environment variables to configure the database connection and other settings. These are documented in the [software package documentation](https://pkg.go.dev/github.com/Defacto2/server). 

There are examples of the environment variables in the [example .env](../init/example.env.local) and the [example .service](../init/defacto2.service) files found in the `init/` directory.

## Source code

The source code requires a local [installation](https://go.dev/doc/install) of [Go version 1.x](https://github.com/Defacto2/server/blob/main/go.mod).

```sh
$ go version

> go version go1.23.4 linux/amd64
```

> [!IMPORTANT]
> While you can compile the application to target Windows environments, it will not function correctly with NTFS file paths. Instead, it is advisable to use Windows Subsystem for Linux.

Clone the source code repository and download the dependencies.

```sh
# clone the repository
git clone https://github.com/Defacto2/server.git

# change to the server repository directory
cd server

# optional, download the dependencies
go mod download
```

Test the application.

```sh
$ go run . --version

> Defacto2 web application version n/a (not a build) for Linux on Intel/AMD 64
```

## Source code tasks

The repository is configured to use the [Task](https://taskfile.dev/installation/) application which needs local installation. The following of tools are expected to be installed on the local machine.

1. [Task](https://taskfile.dev/installation/) is a task runner / build tool.
2. [golangci-lint](https://golangci-lint.run/) is a Go linters aggregator.
3. [GoReleaser](https://goreleaser.com/install/) is a release automation tool for Go projects.

### First time initialization

A new cloned repository needs to download a number of developer specific dependencies.

```sh
# change to the server repository directory
cd server

# run the initialization task
task _init

# confirm the tools are installed
task ver
```

The list of available tasks can be shown.

```sh
$ task # --list-all

task: Available tasks for this project:
* _init:                Initialize this project for the first time after a git clone.
* assets:               Build, compile and compress the web serve CSS and JS assets.
* build:                Build the binary of the web server.
* build-race:           Build the binary of the web server with race detection.
* default:              Task runner for the Defacto2 web server source code.
* doc:                  Generate and browse the application module documentation.
* lint:                 Runs the go formatter and lints the source code.
* lint+:                Runs the deadcode and betteralign linters on the source code.
* nil:                  Run the static analysis techniques to catch Nil dereferences.
* pkg-patch:            Update and apply patches to the web server dependencies.
* pkg-release:          Build the release binary of the web server embeded with the git version tag.  
* pkg-snapshot:         Build the release binary of the web server without a git version tag. 
* pkg-update:           Update the web server dependencies.
* serve-dev:            Run the internal web server in development mode with live reload.
* serve-prod:           Run the internal web server with live reload.
* test:                 Run the test suite.
* testr:                Run the test suite with the slower race detection.
* ver:                  Print the versions of the build and compiler tools.
```

### Configurations

As the application relies on environmental variables for configuration, the Taskfile can use a dot-env file to read in variables for use on tasks.

For example, you can configure various variables while running the `task serve-dev` or `task serve-prod` tasks to point to the downloads, image screenshots, image thumbnails, and a customized database connection URL.

```sh
# change to the server repository directory init directory
cd server/init

# copy the example .env file to the local .env file
cp example.env.local .env.local

# edit the .env file to set the environment variables
nano .env.local
```

An example, the `.env.local` file can be configured as follows.

```ini
# ==============================================================================
#  These are the directory paths serving static files for the artifacts.
#  All directories must be absolute paths and any empty values will disable the
#  relevent feature. For example, an invalid D2_DIR_DOWNLOAD will disable the 
#  artifact downloads.
#  The directories must be readable and writable by the web server.
# ==============================================================================

# List the directory path that holds the named UUID files for the artifact downloads.
D2_DIR_DOWNLOAD=/home/defacto2/downloads

# List the directory path that holds the named UUID files for the artifact images.
D2_DIR_PREVIEW=/home/defacto2/previews

# List the directory path that holds the named UUID files for the artifact thumbnails.
D2_DIR_THUMBNAIL=/home/defacto2/thumbnails

# List the directory path that holds the generated extra files for the artifacts.
D2_DIR_EXTRA=/home/defacto2/extras

# List the directory path that holds the named UUID files that are not linked to
# any database records.
D2_DIR_ORPHANED=/home/defacto2/orphaned

# ==============================================================================
#  These are the PostgreSQL database connection settings.
#  The database is required for accessing and displaying the artifact data.
# ==============================================================================

# The connection string to the PostgreSQL database.
D2_DATABASE_URL=postgres://root:example@localhost:5432/defacto2_ps
```

### Run the development server

Run the internal web server in fast-start, development mode with live reloading of any changes to the Go source code.

```sh
task serve-dev
```

### Run the production server

Run the internal web server in production mode with live reloading of any changes to the Go source code.

```sh
task serve-prod
```
### CSS and JS assets

JavaScript and CSS assets are found in `assets/` and are compiled and compressed into the `public/` directory. 

[ESBuild](https://esbuild.github.io/) is used to compile the JavaScript and it needs to be installed on the local machine. But, ESBuild can be [installed without](https://esbuild.github.io/getting-started/#download-a-build) the need for npm or node.js.

```sh
curl -fsSL https://esbuild.github.io/dl/latest | sh
```

Changes to the assets will require the assets task to be run.

```sh
task assets
```

### Source code linting

The source code is linted using the [golangci-lint](https://golangci-lint.run/) aggregator that runs a number of linters locally.

If you want to optionally lint the CSS and JS assets, you will need to install [Stylelint](https://stylelint.io/) and [ESLint](https://eslint.org/) which will require [node.js](https://nodejs.org/) and a package manager.

```sh
task lint
```

### Testing

The source code has a test suite that can be run.

```sh
task test
```

Or check for race conditions in the test suite.

```sh
task testr
```

### Profiling

The application can be [profiled](https://pkg.go.dev/runtime/pprof) to [check](https://stackademic.com/blog/profiling-go-applications-in-the-right-way-with-examples) for performance bottlenecks or resource usage.

The configuration `D2_LOG_ALL` environment variable must be set to `true` to enable the profiling.

Use the `go tool pprof` to profile the application.

```sh
go tool pprof -http=:8000 -lines -unit=mb https://defacto2.net/debug/pprof/heap
```

- `http=:8000` is the port to view the profile in a local web browser.
- `-lines` will show the source code line numbers in the profile.
- `-unit=mb` will show the memory usage in megabytes.
- `https://defacto2.net/debug/pprof/heap` is the URL to the profile.

Other profiles can be viewed by changing the URL.

The allocs profile, reports statistics as of the total amount of allocated space.
- `https://defacto2.net/debug/pprof/allocs`

Block profile tracks time spent blocked on synchronization primitives, such as sync.Mutex, sync.RWMutex, sync.WaitGroup, sync.Cond, and channel send/receive/select. 
- `https://defacto2.net/debug/pprof/block` for the block profile.
  
The heap profile, it reports statistics as of the most recently completed garbage collection.
- `https://defacto2.net/debug/pprof/heap`
  
The mutex profile tracks contention on mutexes, such as sync.Mutex, sync.RWMutex, and runtime-internal locks.
- `https://defacto2.net/debug/pprof/mutex`

- `https://defacto2.net/debug/pprof/cmdline` for the command line arguments.
- `https://defacto2.net/debug/pprof/goroutine` for the goroutine profile.
- `https://defacto2.net/debug/pprof/profile` for the CPU profile.
- `https://defacto2.net/debug/pprof/symbol` for the symbol profile.
- `https://defacto2.net/debug/pprof/threadcreate` for the threadcreate profile.
- `https://defacto2.net/debug/pprof/trace` for the trace profile.

### Stress testing

The application can be stress tested using the [crawley](https://github.com/s0rg/crawley) tool.

```sh
task build-race
task build-run
crawley -all -brute -depth -1 -headless http://localhost:1323
```

In another terminal, you can view the server's performance using the `top` command, `htop` or view the logs.

```sh
tail -f ~/defacto2_server_info.log
```

### Documentation

The application configuration documentation can be modified in [`doc.go`](../doc.go) and the changes regenerated and [previewed in a web browser](http://localhost:8090).

```sh
task doc
```

```go
// Copyright © 2023-2025 Ben Garrett. All rights reserved.

/*

The [Defacto2] application is a self-contained web server first devised in 2023.
It is built with the Go language and can be easily compiled for significant server 
```

### Building the source code

The source code can be built into a binary for the local machine.

```sh
task build

# or to build with race detection
task buildx
```

The resulting `defacto2-server` binary is built in the repository root directory.

#### Or if you want to build the binary for a [different operating system](https://go.dev/doc/install/source#environment) and architecture.

```sh
# build for macOS on Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o "defacto2-server" server.go

# build for FreeBSD on AMD64
GOOS=freebsd GOARCH=amd64 go build -o "defacto2-server" server.go
```


### Packaging a release

Building the distribution package for the server application is done using a local installation of  [GoReleaser](https://goreleaser.com/install/).

#### Test a package release

To package a snapshot binary for the local machine without a version tag.

```sh
task pkg-snapshot
```

The resulting binary is in the `dist/` directory in the repository root.

#### Packaging a release

```sh
# check the configuration file
goreleaser check --config init/.goreleaser.yaml

# create a new, original tag
git tag -a v1.0.1 -m "First update to the release version."
git push origin v1.0.1

# build the release binary
task pkg-release
```

The resulting built package is found in the `dist/` directory in the repository root.

### Modifying the database schema

The web application relies on an Object-relational mapping (ORM) implementation provided by [SQLBoiler](https://github.com/aarondl/sqlboiler) to simplify development. If the database schema ever changes with a new table column, a modification of an existing column type, or a name change, the ORM code generation requires a rerun.

After modifying the database schema, confirm the local development database connection settings are correct in the SQLBoiler [settings file](../init/.sqlboiler.toml) `init/.sqlboiler.toml`.

```toml
[psql]
schema = "public"
dbname = "defacto2_ps"
host = "localhost"
port = 5432
user = "root"
pass = "example"
sslmode = "disable"

[auto-columns]
created = "createdat"
updated = "updatedat"
deleted = "deletedat"
```

Then in the root of the repository, run go generate.

```sh
go generate
```

The generated code is found in the `internal/postgres/model/` directory and is ready for use.
