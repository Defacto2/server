# Defacto2, <small>web application server</small>

```
      ·      ▒██▀ ▀       ▒██▀ ▀              ▀ ▀▒██             ▀ ▀███ ·
      : ▒██▀ ▓██ ▒██▀▀██▓ ▓██▀▀▒██▀▀███ ▒██▀▀██▓ ▓██▀ ▒██▀▀███ ▒██▀▀▀▀▀ :
 · ··─┼─▓██──███─▓██─▄███─███──▓██──███─▓██──────███──▓██──███─▓██──███─┼─·· ·
      │ ███▄▄██▓ ███▄▄▄▄▄▄██▓  ███▄ ███▄███▄▄███ ███▄▄███▄ ███▄███▄▄███ │
· ··──┼─────────··                defacto2.net               ··─────────┼──·· ·
      │                                                                 :
```

The Defacto2 application is a self-contained web server first devised in 2023.
It is built with the Go language and can be easily compiled for significant server operating systems.
The application relies on a [PostgreSQL](https://www.postgresql.org/) database setup for data queries using a PostgreSQL [database connection](https://www.postgresql.org/docs/current/ecpg-sql-connect.html).

All configurations and modifications to this web application's default settings are through system environment variables.

While you can compile the application to target Windows environments, it is ill-advised as it needs to work correctly with NTFS file paths. Instead, it is advisable to use Windows Subsystem for Linux.

## Download

Currently the application is available as a [standalone binary for Linux](https://github.com/Defacto2/server/releases/download/v0.5.0/df2-server_0.5.0_linux.zip).

## Installation

Installation instructions are provided for [Ubuntu Server](https://ubuntu.com/server) but should be similar for other Linux distributions.

```sh
# change to the home directory
cd ~

# download and unzip the latest release
wget https://github.com/Defacto2/server/releases/latest/download/df2-server_0.5.0_linux.zip
unzip df2-server_0.5.0_linux.zip

# make the binary executable
sudo chmod +x df2-server

# move the binary to the system path
sudo mv df2-server /usr/local/bin

# confirm the binary is executable
df2-server --version
```

## Usage

The web server will run with out any arguments and will be available on the _[localhost](http://localhost)_ with port `1323`.

```sh
df2-server
```

To stop the server, press `CTRL+C`.

## Configuration

The application uses environment variables to configure the database connection and other settings. These are documented in the [software package documentation](https://pkg.go.dev/github.com/Defacto2/server). 

There are examples of the environment variables in the [example .env](../init/example.env.local) and the [example .service](../init/defacto2.service) files found in the `init/` directory.

## Source code

The source code requires a local [installation of Go](https://go.dev/doc/install) version 1.22 or newer. 

> [!IMPORTANT]
> While you can compile the application to target Windows environments, it will not function correctly with NTFS file paths. Instead, it is advisable to use Windows Subsystem for Linux.

Clone the source code repository and download the dependencies.

```sh
# clone the repository
git clone

# change to the server repository directory
cd server

# optional, download the dependencies
go mod download

# test the application
go run . --version
```

## Source code tasks

The repository is configured to use the [Task](https://taskfile.dev/installation/) application which needs local installation.

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
$ task --list-all

task: Available tasks for this project:
* _init:                Initialise this project for the first time after a git clone.
* assets:               Build, compile and compress the web serve CSS and JS assets.
* build:                Build the binary of the web server.
* build-race:           Build the binary of the web server with race detection.
* build-release:        Build the release binary of the web server embeded with the git version tag.  
* build-snapshot:       Build the release binary of the web server without a git version tag. 
* default:              Task runner for the Defacto2 web server source code.
* doc:                  Generate and browse the application module documentation.
* lint:                 Runs the go formatter and lints the source code.
* lint+:                Runs the deadcode and betteralign linters on the source code.
* nil:                  Run the static analysis techniques to catch Nil dereferences.
* pkg-patch:            Update and apply patches to the web server dependencies.
* pkg-update:           Update the web server dependencies.
* serve-dev:            Run the internal web server in development mode with live reload.
* serve-prod:           Run the internal web server with live reload.
* test:                 Run the test suite.
* testr:                Run the test suite with the slower race detection.
* ver:                  Print the versions of the build and compiler tools.
```

### Configuration


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

JavaScript and CSS assets are found in `assets/` and are compiled and compressed into the `public/` directory. Changes to the assets will require the assets task to be run.

```sh
task assets
```

### Source code linting

The source code is linted using the [golangci-lint](https://golangci-lint.run/) aggregator that runs a number of linters locally.

```sh
task lint
```

### Testing

The source code has a test suite that can be run.

```sh
task test
```

### Documentation

The application configuration documentation can be generated and viewed in a web browser.

```sh
task doc
```

Or check for race conditions in the test suite.

```sh
task testr
```

### Building the source code

Building the distribution package for the server application is done using a local installation of  [GoReleaser](https://goreleaser.com/install/).

To build a snapshot binary for the local machine without a version tag.

```sh
task build-snapshot
```

The resulting binary is in the `dist/` directory in the repository root.

### Building a release binary

```sh
# check the configuration file
goreleaser check init/.goreleaser.yaml

# create a new, original tag
git tag -a v1.0.1 -m "First update to the release version."
git push origin v1.0.1

# build the release binary
task build-release
```

The resulting built package is found in the `dist/` directory in the repository root.

