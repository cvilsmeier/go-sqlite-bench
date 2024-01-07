Benchmarks for Golang SQLite Drivers
==============================================================================

This work is sponsored by Monibot - Easy Server and Application Monitoring.
Try out Monibot at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
It's free.


For benchmarks I used the following libraries:

- craw, `crawshaw.io/sqlite`, a CGO-based solution. This is not a `database/sql` driver.

- eaton, `github.com/eatonphil/gosqlite`, a CGO-based solution. This is not a
  `database/sql` driver. (addded by @c4rlo)

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

The benchmark was run on 2024-01-07, with then-current library versions.
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

                     insert      query
    ----------------------------------
    craw            1235 ms     603 ms
    eaton           1286 ms     778 ms
    mattn           1696 ms    1236 ms
    modernc         5615 ms    1236 ms
    ncruces         4082 ms    1370 ms
    sqinn            942 ms     648 ms
    zombie          1941 ms     332 ms


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

                       insert       query
    -------------------------------------
    craw               763 ms      668 ms
    eaton              725 ms      859 ms
    mattn              885 ms     1430 ms
    modernc           3083 ms     1569 ms
    ncruces           2428 ms     1857 ms
    sqinn              576 ms      723 ms
    zombie            1452 ms      502 ms


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

            query/N=10  query/N=100  query/N=1000
    ---------------------------------------------
    craw         15 ms        62 ms        520 ms
    eaton        25 ms        83 ms        670 ms
    mattn        31 ms       124 ms       1093 ms
    modernc      34 ms       130 ms       1096 ms
    ncruces      46 ms       161 ms       1325 ms
    sqinn        37 ms        64 ms        603 ms
    zombie       15 ms        35 ms        213 ms


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

          query/N=50000  query/N=100000  query/N=200000
    ---------------------------------------------------
    craw         197 ms          332 ms          579 ms
    eaton        194 ms          344 ms          665 ms
    mattn        170 ms          301 ms          588 ms
    modernc      279 ms          487 ms          877 ms
    ncruces      233 ms          409 ms          769 ms
    sqinn        556 ms         1082 ms         2273 ms
    zombie       575 ms         1051 ms         2109 ms



### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

            query/N=2  query/N=4  query/N=8
    ---------------------------------------
    craw       700 ms     993 ms    1898 ms
    eaton      816 ms    1264 ms    2264 ms
    mattn     1377 ms    1784 ms    3353 ms
    modernc   2862 ms    7164 ms   18474 ms
    ncruces   1540 ms    1976 ms    3837 ms
    sqinn      843 ms    1383 ms    2152 ms
    zombie     464 ms     624 ms    1121 ms


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Crawshaw and Zombiezen are pretty fast.
- SQLite without CGO is possible.


This work is sponsored by Monibot - Easy Server and Application Monitoring.
Try out Monibot at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
It's free.
