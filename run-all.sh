#!/bin/sh

# failfast
set -e

# must have bin/ directory
stat bin/ > /dev/null

# enter bin/ directory
cd bin/

# run every benchmark twice
date && ./bench-bvinc    bench.db  && ./bench-bvinc    bench.db
date && ./bench-craw     bench.db  && ./bench-craw     bench.db
date && ./bench-eaton    bench.db  && ./bench-eaton    bench.db
date && ./bench-glebarez bench.db  && ./bench-glebarez bench.db
date && ./bench-mattn    bench.db  && ./bench-mattn    bench.db
date && ./bench-modernc  bench.db  && ./bench-modernc  bench.db
date && ./bench-ncruces  bench.db  && ./bench-ncruces  bench.db
date && ./bench-sqinn    bench.db  && ./bench-sqinn    bench.db
date && ./bench-zombie   bench.db  && ./bench-zombie   bench.db
date

# for collecting results, run like this:
#
# $ ./run-all.sh > results/out.txt

