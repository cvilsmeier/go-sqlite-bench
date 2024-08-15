#!/bin/sh

# failfast
set -e

# must be in correct directory
stat go.mod > /dev/null

# build all into _build directory
rm -rf _build
mkdir _build

# build all
echo build craw     && go build -o _build ./cmd/bench-craw
echo build eaton    && go build -o _build ./cmd/bench-eaton
echo build mattn    && go build -o _build ./cmd/bench-mattn
echo build modernc  && go build -o _build ./cmd/bench-modernc
echo build ncruces  && go build -o _build ./cmd/bench-ncruces
echo build sqinn    && go build -o _build ./cmd/bench-sqinn
echo build zombie   && go build -o _build ./cmd/bench-zombie
