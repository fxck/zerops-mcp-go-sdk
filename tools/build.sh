#!/bin/bash

export VERSION="$(git rev-parse --abbrev-ref HEAD):$(git describe --tags --always)-($(git config --get user.name):<$(git config --get user.email)>)"
go build -tags devel \
         -o bin/$1 \
         -gcflags="all=-l -N" \
         -ldflags="all=\"-X=main.version=${VERSION}\"" \
         cmd/mcp-server/main.go