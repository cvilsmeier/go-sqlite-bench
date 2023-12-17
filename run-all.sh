#!/bin/sh

# run every benchmarks twice

date
../bench-craw    ../bench.db
../bench-craw    ../bench.db
date
../bench-mattn   ../bench.db
../bench-mattn   ../bench.db
date
../bench-modernc ../bench.db
../bench-modernc ../bench.db
date
../bench-ncruces ../bench.db
../bench-ncruces ../bench.db
date
../bench-sqinn   ../bench.db
../bench-sqinn   ../bench.db
date
../bench-zombie  ../bench.db
../bench-zombie  ../bench.db
date
