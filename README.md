Benchmarks for Golang SQLite Drivers
==============================================================================

> [!NOTE]
> This work is sponsored by Monibot - Website, Server and Application Monitoring.
> Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).

For benchmarks I used the following libraries:

| Name     | Repository                                                                 | db/sql | cgo   | published  | Remarks |
| :---     | :---                                                                       | :---   | :---  | :---       | :---    |
| bvinc    | [github.com/bvinc/go-sqlite-lite](https://github.com/bvinc/go-sqlite-lite) | no     | yes   | 2019-05-02 | A CGO wrapper library |
| craw     | [github.com/crawshaw/sqlite](https://github.com/crawshaw/sqlite)           | no     | yes   | 2020-06-07 | A CGO wrapper library with statement caching |
| eaton    | [github.com/eatonphil/gosqlite](https://github.com/eatonphil/gosqlite)     | no     | yes   | 2024-08-11 | A CGO wrapper library (addded by @c4rlo) |
| glebarez | [github.com/glebarez/go-sqlite](https://github.com/glebarez/go-sqlite)     | yes    | no    | 2023-12-26 | Based on the modernc libraries (added by @dcarbone) |
| mattn    | [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)         | yes    | yes   | 2026-03-16 | This CGO library is (still) the de-facto standard and widely used |
| modernc  | [modernc.org/sqlite](https://modernc.org/sqlite)                           | yes    | no    | 2026-03-17 | A pure Go solution, based on the SQLite C code transpiled to Go |
| ncruces  | [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3)     | yes    | no    | 2026-03-21 | A pure Go solution based on a WASM build of SQLite |
| sqinn    | [github.com/cvilsmeier/sqinn-go](https://github.com/cvilsmeier/sqinn-go)   | no     | no    | 2026-03-18 | A solution without CGO. It uses [github.com/cvilsmeier/sqinn](https://github.com/cvilsmeier/sqinn) to access SQLite database files. |
| zombie   | [github.com/zombiezen/go-sqlite](https://github.com/zombiezen/go-sqlite)   | no     | no    | 2025-05-23 | A pure-Go rewrite of the crawshaw driver, using the modernc libraries. |

The test setup is as follows:

- OS: Debian/GNU Linux amd64 version 12.13
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 cores
- RAM: 32GB
- Disk: 1TB NVME SSD
- go version go1.26.0 linux/amd64

The benchmark was run on 2026-03-23, with current library versions,
see go.mod file. Each test was run twice. The better result was then
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
    bvinc;         1021;    438;
    craw;          1035;    429;
    eaton;         1010;    511;
    glebarez;      5162;    749;
    mattn;         1480;    871;
    modernc;       2419;    758;
    ncruces;       2719;    850;
    sqinn;          645;    242;
    zombie;        1746;    263;


### Real

Insert 100 user with 20 articles per user and 20 comments per article.
Each user is inserted in a separate transaction.
Then query each user by email, and left-join articles and comments.
This benchmark is used to simulate real-world use cases.

![](results/real.png)

    Real;      insert;  query;
    bvinc;       1230;     61;
    craw;        1258;     45;
    eaton;       1188;     70;
    glebarez;    1815;    129;
    mattn;       1586;    104;
    modernc;     1759;    130;
    ncruces;     1469;    129;
    sqinn;       1239;     51;
    zombie;      1892;     59;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    bvinc;          672;    558;
    craw;           647;    476;
    eaton;          646;    653;
    glebarez;      2814;   1075;
    mattn;          812;    998;
    modernc;       1554;   1068;
    ncruces;       1749;   1263;
    sqinn;          475;    258;
    zombie;        1270;    488;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    bvinc;               23;          56;          415;
    craw;                12;          46;          380;
    eaton;               22;          62;          451;
    glebarez;            21;          86;          702;
    mattn;               20;          96;          819;
    modernc;             31;          87;          692;
    ncruces;             33;         103;          837;
    sqinn;               29;          51;          320;
    zombie;              16;          34;          267;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    bvinc;                 178;            289;            531;
    craw;                  187;            304;            535;
    eaton;                 133;            226;            499;
    glebarez;              406;            717;           1129;
    mattn;                 122;            207;            376;
    modernc;               401;            629;           1094;
    ncruces;               151;            287;            528;
    sqinn;                 285;            544;           1132;
    zombie;                329;            561;            952;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    bvinc;             508;       795;      1526;
    craw;              514;       783;      1398;
    eaton;             633;       887;      1637;
    glebarez;          803;      1192;      2105;
    mattn;             948;      1237;      2395;
    modernc;           850;      1190;      2061;
    ncruces;           981;      1261;      2355;
    sqinn;             394;       700;      1273;
    zombie;            311;       556;      1033;


Summary
------------------------------------------------------------------------------

- We cannot declare a clear winner, it all depends on the use case.
- SQLite without CGO is possible nowadays.
