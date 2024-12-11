Benchmarks for Golang SQLite Drivers
==============================================================================

> [!NOTE]
> This work is sponsored by Monibot - Website, Server and Application Monitoring.
> Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).


For benchmarks I used the following libraries:

- craw, `crawshaw.io/sqlite`, a CGO-based solution. This is not a `database/sql` driver.

- eaton, `github.com/eatonphil/gosqlite`, a CGO-based solution. This is not a
    `database/sql` driver. (addded by @c4rlo)

- glebarez, `github.com/glebarez/go-sqlite`, a pure Go solution. This is a newer library,
    based on the SQLite C code re-written in Go (added by @dcarbone).

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

- OS: Debian/GNU Linux amd64 version 12.8
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 cores
- RAM: 32GB
- Disk: 1TB NVME SSD
- go version go1.23.4 linux/amd64

The benchmark was run on 2024-12-11, with current library versions,
see go.mod file. Each test was run once for warmup. The second run was then
recorded. This is not very scientific.


A general note on benchmarks and this repository:
------------------------------------------------------------------------------

Do not trust benchmarks, write your own. This specific benchmark is modelled
after my very own database usage scenarios. Your scenarios may be totally
different.

This is also the reason this repository is open-source, but not open-contribution.
There are many good ideas to improve this benchmark: More driver libs, different
lib versions, windows, macOS, different SQLite journal- and sync modes, etc.

Unfortunately, supporting all this would take too much time for me.
I read all proposals but have to, unfortunately, be very selective as to what
to include in this project.



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


Benchmark Results
------------------------------------------------------------------------------

Result times are measured in milliseconds. Lower numbers indicate better
performance.


### Simple

Insert 1 million user rows in one database transaction.
Then query all users once.

![](results/simple.png)

    Simple;      insert;  query;
    craw;          1304;    587;
    eaton;         1119;    702;
    glebarez;      5700;   1207;
    mattn;         1682;   1187;
    modernc;       5508;   1178;
    ncruces;       3244;   1053;
    sqinn;          966;    623;
    zombie;        1831;    314;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    craw;           754;    593;
    eaton;          727;    816;
    glebarez;      3255;   1495;
    mattn;          923;   1261;
    modernc;       3219;   1497;
    ncruces;       1948;   1284;
    sqinn;          646;    729;
    zombie;        1449;    501;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    craw;                15;          66;          479;
    eaton;               25;          75;          597;
    glebarez;            24;         127;         1041;
    mattn;               19;         115;          993;
    modernc;             34;         126;         1048;
    ncruces;             43;         124;         1057;
    sqinn;               20;          66;          702;
    zombie;              18;          37;          273;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    craw;                  208;            365;            714;
    eaton;                 184;            325;            609;
    glebarez;              244;            723;           1146;
    mattn;                 154;            284;            501;
    modernc;               232;            653;           1188;
    ncruces;               212;            414;            790;
    sqinn;                 588;           1114;           2103;
    zombie;                180;            558;           1037;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    craw;              592;       957;      1629;
    eaton;             761;      1145;      1962;
    glebarez;         2698;      7078;     18088;
    mattn;            1298;      1725;      2915;
    modernc;          2606;      7044;     17837;
    ncruces;          1153;      1527;      2614;
    sqinn;             634;      1370;      2333;
    zombie;            399;       625;      1082;


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Crawshaw and Zombiezen are pretty fast.
- SQLite without CGO is possible.


> [!NOTE]
> This work is sponsored by Monibot - Website, Server and Application Monitoring.
> Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
