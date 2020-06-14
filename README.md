Benchmarks for Sqinn-Go
==============================================================================

Performance tests show that Sqinn-Go performance is roughly the same as cgo
solutions, sometimes even better.

For benchmarks I used `github.com/mattn/go-sqlite3` ("mattn")
and `crawshaw.io/sqlite` ("craw"). Both are cgo-based solutions.

The test setup is as follows:

- OS: Windows 10 Home x64 Version 1909 Build 18363
- CPU: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz, 2592 MHz, 4 Cores
- RAM: 16GB
- Disk: 256GB SSD

## Benchmark 1: Simple

Inserting and querying 1 million rows in one goroutine. The schema is only one
table

	CREATE TABLE users (
		id INTEGER PRIMARY KEY NOT NULL,
		name VARCHAR,
		age INTEGER,
		rating REAL
	);

All inserts are done in one transaction. The results are (lower is better):

                       mattn    craw    sqinn
    simple/insert      2.8 s    2.1 s   1.5 s
    simple/query       2.3 s    1.3 s   1.3 s


## Benchmark 2: Complex Schema

A more complex table schema with foreign key constraints and many indices.
Inserting and querying 200000 rows in one goroutine.

All inserts are done in one transaction. The results are (lower is better):

                       mattn    craw    sqinn
    complex/insert     2.0 s    1.8 s   1.7 s
    complex/query      1.4 s    1.1 s   1.3 s


## Benchmark 3: Concurrent

Querying a table with 1 million rows concurrently. Spin up N goroutines, where
each goroutine queries all 1000000 rows.

All inserts are done in one transaction. The results are (lower is better):

                       mattn    craw    sqinn
    concur/N=2         1.3 s    0.9 s   0.9 s
    concur/N=4         1.5 s    1.0 s   1.2 s
    concur/N=8         2.3 s    1.6 s   2.0 s


## Summary

The crawshaw library is faster than the mattn driver. Sqinn performance about
as well as the crawshaw library.

But: Every application is different, and I recommend that you perform
benchmarks based on the typical workload of your application. As always, it
depends.
