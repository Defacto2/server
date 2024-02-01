# Defacto2, the web application server

```
      ·      ▒██▀ ▀       ▒██▀ ▀              ▀ ▀▒██             ▀ ▀███ ·
      : ▒██▀ ▓██ ▒██▀▀██▓ ▓██▀▀▒██▀▀███ ▒██▀▀██▓ ▓██▀ ▒██▀▀███ ▒██▀▀▀▀▀ :
 · ··─┼─▓██──███─▓██─▄███─███──▓██──███─▓██──────███──▓██──███─▓██──███─┼─·· ·
      │ ███▄▄██▓ ███▄▄▄▄▄▄██▓  ███▄ ███▄███▄▄███ ███▄▄███▄ ███▄███▄▄███ │
· ··──┼─────────··                defacto2.net               ··─────────┼──·· ·
      │                                                                 :
```

The [Defacto2](https://defacto2.net) web server is a self-contained application, first created in 2023 and built with the [Go language](https://go.dev/). And can be easily compiled for [major operating systems](https://pkg.go.dev/internal/platform#pkg-variables).

The web server relies on a [PostgreSQL database](https://www.postgresql.org/) for data queries, best provided using a container such as [Docker](https://www.docker.com/).

All configurations and settings for the web application are through system environment variables.
Variables are handled within the container's environment on a production setup, such as with a Docker container.

## Download

Numerous downloads are available for [Windows](https://github.com/Defacto2/server/releases/latest/download/defacto2-app_windows_amd64_v1.zip), [macOS](https://github.com/Defacto2/server/releases/latest/download/df2-server_darwin_all.zip), [Linux](https://github.com/Defacto2/server/releases/latest/download/defacto2-app_linux_amd64_v1.zip.zip) and more.

The server app is a standalone, self-contained terminal program, but requires additional setups such as an running [Defacto2 PostgreSQL database](https://github.com/Defacto2/database-ps).

## Installation

All the instructions assume macOS, Linux or Windows Subsystem for Linux (WSL).

### Docker

The recommended way to run the server app is to use a [Docker](https://www.docker.com/) container. 

#### Database

Firstly, set up the [Defacto2 PostgreSQL database](https://github.com/Defacto2/database-ps).

```sh
# clone the database repository
cd ~
git clone git@github.com:Defacto2/database-ps.git
cd ~/database-ps

# migrate the Defacto2 data from MySQL to PostgreSQL
docker compose --profile migrater up

# stop the running database by pressing CTRL+C
# cleanup the unnecessary volumes and containers
docker compose rm migrate mysql dbdump --stop
docker volume rm database-ps_tmpdump database-ps_tmpsql

# restart the database to run in the background
docker compose up -d
```

#### Web server

A preconfigured docker-compose file exists for use with Docker Desktop or docker.

[Download the `docker-compose.yml` file](https://github.com/Defacto2/server/blob/main/docker-compose.yml) to a local directory such as `~/df2-server`.

```sh
# create the local directory
mkdir ~/df2-server

# copy the downloaded docker-compose.yml file to the directory
cp ~/downloads/docker-compose.yml ~/df2-server
```

Create a `.env` file to store our environment variables for the container and copy [the .env example](#example-env) content and save.

```sh
cd ~/df2-server

# create the .env file and paste then save the example content
touch .env
nano .env
```

Start the container and the web server will be available on the _localhost_ with port `1323`.

#### http://localhost:1323

```sh
docker compose up -d
```

### Example `.env` 

Docker uses the `.env` file to set container environment variables.

```ini
# ===================
#  Database settings
# ===================

# Database username used to connect
PS_USERNAME=root

# Password for the database username
PS_PASSWORD=example

# ===================
#  Optional, directory paths for the serving of static files.
# ===================

# The absolute directory path that holds the UUID named files for the downloads.
D2_DOWNLOAD_DIR=/home/ben/defacto2/downloads

# The absolute directory path that holds the UUID named files for the images.
D2_PREVIEW_DIR=/home/ben/defacto2/images

# The absolute directory path that holds the UUID named files for the thumbnails.
D2_THUMBNAIL_DIR=/home/ben/defacto2/thumbnails

# ===================
#  Web application and server settings
# ===================
#
# The unencrypted port number that the HTTP web server will listen on.
D2_HTTP_PORT=1323
```

### Local

Download the latest release for your operating system from the [releases page](https://github.com/Defacto2/server/releases).

Uncompress the downloaded file and run the binary. The application uses environment variables to configure the database connection and other settings. But these can be set and unset using a `.env` file and a shell script.

```sh
# create the local directory for the binary and configuration
mkdir ~/df2-server

# uncompressed the downloaded file to the directory
unzip ~/downloads/defacto2-app_linux_amd64_v1.zip -d ~/df2-server
cd ~/df2-server

# confirm the binary is executable
chmod +x df2-server
./df2-server --version

# create the .env file and edit, paste and save the example content
touch .env
nano .env

# create, paste and save the the shell script example content (listed below) to a file named run.sh
touch run.sh
nano run.sh

# make the shell script executable and run it
chmod +x run.sh
./run.sh
```

#### Example `run.sh` shell script

```sh
#!/bin/bash

# The following script is used to run the server with environment variables.
# The environment variables are loaded from a file named ".env" but 
# this can be changed by modifying the FILENAME variable below.
#
# The df2-server binary should be in the same directory as this script.

# Filename containing the environment variables
FILENAME=.env

# Load environment variables from .env
echo -e "Loading environment variables from $FILENAME\n"
export $(grep -E -v '^#' $FILENAME | xargs)

# Run the server
./df2-server

# Unset environment variables from .env
echo -e "\nUnset environment variables from $FILENAME\n"
unset $(grep -E -v '^#' $FILENAME | sed -E 's/(.*)=.*/\1/' | xargs)

```

## Usage

The web application has a basic help for the command line interface.

```sh
./df2-server --help
```

More detailed information is available in the [package documentation](https://pkg.go.dev/github.com/Defacto2/server).

### Source code

Instructions for editing, testing and running the source code are available in the [package documentation](https://pkg.go.dev/github.com/Defacto2/server).
