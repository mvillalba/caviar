#!/usr/bin/env bash
set -e

go build serve.go
cavundle serve assets
