#!/usr/bin/env bash

set -e

mkdir -p $GOPATH/src/github.com/mvillalba
cp -R /home/martin/wks/gopath/src/github.com/mvillalba/caviar $GOPATH/src/github.com/mvillalba

CAV=$GOPATH/src/github.com/mvillalba/caviar/caviarize.sh

go get -v bitbucket.org/kardianos/osext
$CAV github.com/revel/revel
$CAV code.google.com/p/go.net
$CAV code.google.com/p/go.net/websocket
$CAV github.com/agtorre/gocolorize
$CAV github.com/howeyc/fsnotify
$CAV github.com/robfig/config
$CAV github.com/robfig/pathtree
$CAV github.com/streadway/simpleuuid
