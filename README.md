Benchmarks for Sqinn-Go
==============================================================================

Performance tests show that Sqinn-Go is comparable to cgo-based solutions,
depending on the use case and on the cgo-solution that's used.

For benchmarks I used the following libraries:

- `github.com/mattn/go-sqlite3`, a cgo-based solution. This library is
  (still) the de-facto standard and widely used. 

- `modernc.org/sqlite`, a pure-go solution. This is a newer library,
  based purely on a go implementation of the sqlite3 source code.

- `crawshaw.io/sqlite`, a cgo-based solution. This is not a `database/sql` driver.

The test setup is as follows:

- OS: Debian/GNU Linux amd64 version 11.2
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 Cores
- RAM: 16GB
- Disk: 1TB NVME SSD
- go version: go1.17.8 linux/amd64

Result times are measured in milliseconds. Lower numbers indicate better
performance. Each benchmark was executed twice (for warmup) and the second
run was taken.

For details, please see the souce code. The next section describes the
benchmarks, then follows a summary.


Benchmarks
------------------------------------------------------------------------------

### Simple

Insert 1 million rows in a simple table. Then query all rows.
All inserts are done in one transaction.

                       mattn   cznic  crawshaw   sqinn
    simple/insert       1780    4610      1245     729
    simple/query        1350    1035       748     549


### Complex

A more complex table schema with foreign key constraints and many indices.
Inserting and querying 200000 rows in one goroutine.
All inserts are done in one transaction.

                       mattn   cznic  crawshaw   sqinn
    complex/insert       954    3068       868     834
    complex/query        800    1032       629     555


### Many

Querying a simple table with N rows 1000 times in one goroutine.
This benchmark is used to simulate the "N+1 Select" problem. N is the number
of rows in the table.

                       mattn   cznic  crawshaw   sqinn
    many/N=10             18      32        17      19
    many/N=100           106      97        86      52
    many/N=1000         1044     898       684     425


### Large

Querying a table with very large row contents. N is the size of the row in
bytes.

                       mattn   cznic  crawshaw   sqinn
    large/N=2000          70      69        84      82
    large/N=4000         114     258        95     215
    large/N=8000         240     312       148     439


### Concurrent

Querying a table with 1 million rows concurrently. Spin up N goroutines, where
each goroutine queries all 1000000 rows.

                       mattn   cznic  crawshaw   sqinn
    concurrent/N=2       953     724       581     388
    concurrent/N=4      1217     936       789     575
    concurrent/N=8      1860    2011      1250    1131


Summary
------------------------------------------------------------------------------

In benchmark "Large", Sqinn-Go is slower than the other solutions:
When dealing with very large rows, Sqinn-Go has to shuffle a lot of data
across process boundaries, and that takes time.

In all other benchmarks, Sqinn-Go is the fastest solution.
