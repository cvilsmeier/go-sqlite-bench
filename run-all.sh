#!/bin/sh

# run all benchmarks one-by-one

date
../bench-craw    ../bench.db
../bench-mattn   ../bench.db
../bench-modernc ../bench.db
../bench-ncruces ../bench.db
../bench-sqinn   ../bench.db
../bench-zombie  ../bench.db
date
