#!/bin/sh

# failfast
set -e

# must be in correct directory
stat go.mod > /dev/null

# build all into bin/ directory
rm -rf bin/
mkdir bin/

# build all
echo build bvinc    && go build -o bin/ ./cmd/bench-bvinc
echo build craw     && go build -o bin/ ./cmd/bench-craw
echo build eaton    && go build -o bin/ ./cmd/bench-eaton
echo build glebarez && go build -o bin/ ./cmd/bench-glebarez
echo build mattn    && go build -o bin/ ./cmd/bench-mattn
echo build modernc  && go build -o bin/ ./cmd/bench-modernc
echo build ncruces  && go build -o bin/ ./cmd/bench-ncruces
echo build sqinn    && go build -o bin/ ./cmd/bench-sqinn
echo build zombie   && go build -o bin/ ./cmd/bench-zombie
