#!/bin/sh

# build all benchmarks in parent directory

go build -o .. ./cmd/bench-craw
go build -o .. ./cmd/bench-eaton
go build -o .. ./cmd/bench-mattn
go build -o .. ./cmd/bench-modernc
go build -o .. ./cmd/bench-ncruces
go build -o .. ./cmd/bench-sqinn
go build -o .. ./cmd/bench-zombie
