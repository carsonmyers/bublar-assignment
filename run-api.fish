#!/usr/bin/env fish
set -l pkg (go env GOPATH)/src/github.com/carsonmyers/bublar-assignment
set -l build $pkg/cmd/api

cd $build

env LOG_LEVEL=debug \
API_HOST=0.0.0.0 API_PORT=62880 API_PROTOCOL=http \
LOCATIONS_HOST=localhost LOCATIONS_PORT=49800 \
PLAYERS_HOST=localhost PLAYERS_PORT=49801 \
    go run .

cd -