Benchmarks for Golang SQLite Drivers
==============================================================================

This work is sponsored by Monibot - Easy Server and Application Monitoring.
Try out Monibot at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
It's free.


For benchmarks I used the following libraries:

- craw, `crawshaw.io/sqlite`, a CGO-based solution. This is not a `database/sql` driver.

- mattn, `github.com/mattn/go-sqlite3`, a CGO-based solution. This library is
  (still) the de-facto standard and widely used. 

- modernc, `modernc.org/sqlite`, a pure Go solution. This is a newer library,
  based on the SQLite C code re-written in Go.

- ncruces, `github.com/ncruces/go-sqlite3`, a pure Go solution based on WASM (?). 

- sqinn, `github.com/cvilsmeier/sqinn-go`, a solution without CGO. It uses
  `github.com/cvilsmeier/sqinn` to access SQLite database files.

- zombie, `github.com/zombiezen/go-sqlite`, a rewrite of the crawshaw driver, using the
  modernc libraries. This is not a `database/sql` driver.


The test setup is as follows:

- OS: Debian/GNU Linux amd64 version 12.4
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 4 physical cores, 8 logical cores
- RAM: 16GB
- Disk: 1TB NVME SSD
- go version go1.21.5 linux/amd64

The benchmark was run on 2023-12-17, with then-current library versions.
See go.mod for library versions. Each test was run once for warmup.
The second run was then recorded. This is not very scientific.


A general note on benchmarks:

Do not trust benchmarks, write your own. This specific benchmark is modelled
after my very own database usage scenarios. Your scenarios may be totally
different.


Database Schema
------------------------------------------------------------------------------

The test database consist of the following tables and indizes:

    PRAGMA journal_mode=DELETE;
    PRAGMA synchronous=FULL;
    PRAGMA foreign_keys=1;
    PRAGMA busy_timeout=5000;

    CREATE TABLE users (
        id INTEGER PRIMARY KEY NOT NULL,
        created INTEGER NOT NULL,
        email TEXT NOT NULL,
        active INTEGER NOT NULL);
    CREATE INDEX users_created ON users(created);

    CREATE TABLE articles (
        id INTEGER PRIMARY KEY NOT NULL,
        created INTEGER NOT NULL,  
        userId INTEGER NOT NULL REFERENCES users(id),
        text TEXT NOT NULL);
    CREATE INDEX articles_created ON articles(created);
    CREATE INDEX articles_userId ON articles(userId);

    CREATE TABLE comments (
        id INTEGER PRIMARY KEY NOT NULL,
        created INTEGER NOT NULL,
        articleId INTEGER NOT NULL REFERENCES articles(id),
        text TEXT NOT NULL);
    CREATE INDEX comments_created ON comments(created);
    CREATE INDEX comments_articleId ON comments(articleId);


Benchmarks
------------------------------------------------------------------------------

Result times are measured in milliseconds. Lower numbers indicate better
performance.


### Simple

Insert 1 million user rows in one database transaction.
Then query all users once.

![](results/simple.png)

                      insert        query
    -------------------------------------
    craw             1209 ms       594 ms
    mattn            1733 ms      1352 ms
    modernc          5554 ms      1300 ms
    ncruces          4144 ms      1446 ms
    sqinn             886 ms       638 ms
    zombie           1969 ms       345 ms


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

                       insert       query
    -------------------------------------
    craw               742 ms      648 ms
    mattn              927 ms     1401 ms
    modernc           3088 ms     1685 ms
    ncruces           2406 ms     1901 ms
    sqinn              571 ms      710 ms
    zombie            1437 ms      504 ms


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

            query/N=10  query/N=100  query/N=1000
    ---------------------------------------------
    craw         15 ms        63 ms        519 ms
    mattn        31 ms       131 ms       1172 ms
    modernc      22 ms       129 ms       1170 ms
    ncruces      42 ms       166 ms       1368 ms
    sqinn        20 ms        70 ms        587 ms
    zombie       17 ms        35 ms        211 ms


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

          query/N=50000  query/N=100000  query/N=200000
    ---------------------------------------------------
    craw         193 ms          348 ms          597 ms
    mattn        167 ms          303 ms          524 ms
    modernc      276 ms          471 ms          836 ms
    ncruces      207 ms          351 ms          733 ms
    sqinn        519 ms         1077 ms         2300 ms
    zombie       576 ms         1109 ms         2170 ms


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

            query/N=2  query/N=4  query/N=8
    ---------------------------------------
    craw       770 ms    1028 ms    1907 ms
    mattn     1460 ms    1852 ms    3435 ms
    modernc   2843 ms    7036 ms   18272 ms
    ncruces   1541 ms    1883 ms    3925 ms
    sqinn      832 ms    1292 ms    2430 ms
    zombie     470 ms     637 ms    1118 ms


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Crawshaw and Zombiezen are pretty fast.
- Mattn, although the de-facto standard, is not the best overall solution.
- SQLite without CGO is possible.


This work is sponsored by Monibot - Easy Server and Application Monitoring.
Try out Monibot at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
It's free.
