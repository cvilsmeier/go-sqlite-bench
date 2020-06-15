Benchmarks for Sqinn-Go
==============================================================================

Performance tests show that Sqinn-Go is comparable to cgo-based solutions,
depending on the use case and on the cgo-solution that's used.

For benchmarks I used `github.com/mattn/go-sqlite3` and `crawshaw.io/sqlite`,
two cgo-based solutions. The first (mattn) is the de-facto standard and widely
used. The second (crawshaw) is a newer library. Most notabley, mattn is a
`database/sql` driver, crawshaw is not.

The test setup is as follows:

- OS: Windows 10 Home x64 Version 1909 Build 18363
- CPU: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz, 2592 MHz, 4 Cores
- RAM: 16GB
- Disk: 256GB SSD

Result times are measured in milliseconds. Lower numbers indicate better
performance. Each benchmark was executed twice and the minimum time of each
run was taken.

For details, please see the souce code. The next section describes the
benchmarks, then follows a summary.


Benchmarks
------------------------------------------------------------------------------

### Simple

Insert 1 million rows in a simple table. Then query all rows.
All inserts are done in one transaction.

                       mattn  crawshaw     sqinn
    simple/insert       2901      2140      1563
    simple/query        2239      1287      1390


### Complex

A more complex table schema with foreign key constraints and many indices.
Inserting and querying 200000 rows in one goroutine.

All inserts are done in one transaction.

                       mattn  crawshaw     sqinn
    complex/insert      2066      1817      1683
    complex/query       1458      1129      1338


### Many

Querying a simple table with N rows 1000 times in one goroutine.
This benchmark is used to simulate the "N+1 Select" problem. N is the number
of rows in the table.

                       mattn  crawshaw     sqinn
    many/N=10             97        78       134
    many/N=100           246       194       276
    many/N=1000         1797      1240      1436


### Large

Querying a table with very large row contents. N is the size of the row in
bytes.

                       mattn  crawshaw     sqinn
    large/N=2000         119        87       341
    large/N=4000         361       322       760
    large/N=8000         701       650      1531


### Concurrent

Querying a table with 1 million rows concurrently. Spin up N goroutines, where
each goroutine queries all 1000000 rows.

                       mattn  crawshaw     sqinn
    concurrent/N=2      1332       865       951    
    concurrent/N=4      1505       989      1207    
    concurrent/N=8      2347      1557      2044     


Summary
------------------------------------------------------------------------------

In all of the above benchmarks, the crawshaw library is faster than the mattn
library.

In most benchmarks, Sqinn-Go lies between mattn and crawshaw, with two
exceptions.

- Benchmark "Large": When dealing with very large rows, Sqinn-Go
  has to shuffle a lot of data across process boundaries, and that takes time.

- Benchmark "Many": When handling many fast-executing queries, Sqinn-Go has to
  make a lot of process switches, which is time consuming.

Every application is different, and I recommend that you perform
benchmarks based on the typical workload of your application. As always, it
depends.
