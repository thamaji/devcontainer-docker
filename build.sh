#!/bin/bash
set -eu

go build -ldflags="-s -w" -trimpath -o docker
