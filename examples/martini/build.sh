#!/usr/bin/env bash

set -e

go get -v github.com/mvillalba/caviar
CAV=$GOPATH/src/github.com/mvillalba/caviar/caviarize.sh

$CAV github.com/codegangsta/martini
go build martini.go
cavundle --debug martini assets
