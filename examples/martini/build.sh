#!/usr/bin/env bash

set -e

export GOPATH=/tmp/caviar_gopath
rm -rf $GOPATH

caviarize github.com/codegangsta/martini
go build martini.go
cavundle martini assets
