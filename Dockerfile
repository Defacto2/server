# syntax=docker/dockerfile:1
#
# Simple instructions to build and run the server
#
# Build:
#   docker build -t defacto2-app .
#
# Run:
#   docker run -p 1323:1323 -it --rm --name defacto2-server defacto2-app


# Build stage
# FROM golang:1.21 AS build
# WORKDIR /app

# # Copy go.mod and go.sum files and download dependencies
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy source code and build binary
# COPY . .

# RUN go build -o server server.go

# RUN mkdir /root/.config

# EXPOSE 1323

# CMD ["./server"]


FROM alpine:latest

RUN mkdir /root/.config

COPY dist/server_linux_amd64_v1/df2-server /usr/local/bin/df2-server

EXPOSE 1323

ENTRYPOINT ["df2-server"]