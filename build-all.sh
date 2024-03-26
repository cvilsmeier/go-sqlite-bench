#!/bin/sh

# build all benchmarks in parent directory

echo build craw     && go build -o .. ./cmd/bench-craw
echo build eaton    && go build -o .. ./cmd/bench-eaton
echo build mattn    && go build -o .. ./cmd/bench-mattn
echo build modernc  && go build -o .. ./cmd/bench-modernc
echo build ncruces  && go build -o .. ./cmd/bench-ncruces
echo build sqinn    && go build -o .. ./cmd/bench-sqinn
echo build zombie   && go build -o .. ./cmd/bench-zombie
