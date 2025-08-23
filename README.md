Benchmarks for Golang SQLite Drivers
==============================================================================

> [!NOTE]
> This work is sponsored by Monibot - Website, Server and Application Monitoring.
> Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).


For benchmarks I used the following libraries:

- bvinc, [github.com/bvinc/go-sqlite-lite](https://github.com/bvinc/go-sqlite-lite),
  a CGO-based solution.
  This is not a `database/sql` driver.

- craw, [github.com/crawshaw/sqlite](https://github.com/crawshaw/sqlite),
  a CGO-based solution.
  This is not a `database/sql` driver.

- eaton, [github.com/eatonphil/gosqlite](https://github.com/eatonphil/gosqlite),
  a CGO-based solution.
  This is not a `database/sql` driver. (addded by @c4rlo)

- glebarez, [github.com/glebarez/go-sqlite](https://github.com/glebarez/go-sqlite),
  a pure Go solution. This is a newer library, based on the modernc libraries (added by @dcarbone).
  This is a `database/sql` driver.

- mattn, [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3),
  a CGO-based solution. This library is (still) the de-facto standard and widely used. 
  This is a `database/sql` driver.

- modernc, [modernc.org/sqlite](https://modernc.org/sqlite),
  a pure Go solution. This is a newer library, based on the SQLite C code transpiled to Go.
  This is a `database/sql` driver.

- ncruces, [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3),
  a pure Go solution based on SQLite's WASM build and wazero. 
  This is a `database/sql` driver.

- sqinn, [github.com/cvilsmeier/sqinn-go](https://github.com/cvilsmeier/sqinn-go),
  a solution without CGO. It uses [github.com/cvilsmeier/sqinn](https://github.com/cvilsmeier/sqinn)
  to access SQLite database files.
  This is not a `database/sql` driver.

- zombie, [github.com/zombiezen/go-sqlite](https://github.com/zombiezen/go-sqlite),
  a rewrite of the crawshaw driver, using the modernc libraries.
  This is not a `database/sql` driver.


For better overview, here is a table:

| Name     | Repository                                                                 | db/sql | cgo  | Remarks |
| :---     | :---                                                                       | :---   | :--- | :---    |
| bvinc    | [github.com/bvinc/go-sqlite-lite](https://github.com/bvinc/go-sqlite-lite) | -      | yes  | |
| craw     | [github.com/crawshaw/sqlite](https://github.com/crawshaw/sqlite)           | -      | yes  | |
| eaton    | [github.com/eatonphil/gosqlite](https://github.com/eatonphil/gosqlite)     | -      | yes  | (addded by @c4rlo) |
| glebarez | [github.com/glebarez/go-sqlite](https://github.com/glebarez/go-sqlite)     | yes    | -    | This is a newer library, based on the modernc libraries (added by @dcarbone) |
| mattn    | [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)         | yes    | yes  | This library is (still) the de-facto standard and widely used. |
| modernc  | [modernc.org/sqlite](https://modernc.org/sqlite)                           | yes    | -    | This is a newer library, based on the SQLite C code transpiled to Go. |
| ncruces  | [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3)     | yes    | -    | This library is based on SQLite's WASM build and wazero. |
| sqinn    | [github.com/cvilsmeier/sqinn-go](https://github.com/cvilsmeier/sqinn-go)   | -      | -    | It uses [github.com/cvilsmeier/sqinn](https://github.com/cvilsmeier/sqinn) to access SQLite database files.|
| zombie   | [github.com/zombiezen/go-sqlite](https://github.com/zombiezen/go-sqlite)   | -      | -    | A rewrite of the crawshaw driver, using the modernc libraries. |


The test setup is as follows:

- OS: Debian/GNU Linux amd64 version 12.11
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 cores
- RAM: 32GB
- Disk: 1TB NVME SSD
- go version go1.24.5 linux/amd64

The benchmark was run on 2025-08-17, with current library versions,
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
    bvinc;         1140;    574;
    craw;          1173;    501;
    eaton;         1107;    631;
    glebarez;      5131;    761;
    mattn;         1531;   1018;
    modernc;       5288;    760;
    ncruces;       3046;    910;
    sqinn;          717;    247;
    zombie;        1791;    264;


### Real

Insert 100 user with 20 articles per user and 20 comments per article.
Each user is inserted in a separate transaction.
Then query each user by email, and left-join articles and comments.
This benchmark is used to simulate real-world use cases.

![](results/real.png)

    Real;      insert;  query;
    bvinc;       1166;     75;
    craw;        1212;     58;
    eaton;       1107;     80;
    glebarez;    1673;    134;
    mattn;       1416;    120;
    modernc;     1641;    130;
    ncruces;     1364;    127;
    sqinn;       1236;     48;
    zombie;      1762;     59;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    bvinc;          701;    692;
    craw;           699;    595;
    eaton;          686;    756;
    glebarez;      2918;   1093;
    mattn;          843;   1187;
    modernc;       2909;   1100;
    ncruces;       1834;   1211;
    sqinn;          457;    292;
    zombie;        1286;    481;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    bvinc;               15;          66;          532;
    craw;                13;          59;          499;
    eaton;               24;          77;          584;
    glebarez;            29;          90;          723;
    mattn;               30;         111;         1000;
    modernc;             30;          91;          736;
    ncruces;             40;         116;          965;
    sqinn;               40;          58;          339;
    zombie;              15;          34;          218;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    bvinc;                 149;            303;            560;
    craw;                  167;            296;            574;
    eaton;                 138;            227;            450;
    glebarez;              404;            664;           1160;
    mattn;                 131;            238;            449;
    modernc;               387;            674;           1116;
    ncruces;               171;            341;            648;
    sqinn;                 292;            507;           1074;
    zombie;                335;            591;           1000;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    bvinc;             686;       915;      1729;
    craw;              586;       869;      1633;
    eaton;             714;      1015;      1924;
    glebarez;          856;      1182;      2177;
    mattn;            1149;      1478;      2830;
    modernc;           870;      1170;      2139;
    ncruces;           982;      1300;      2516;
    sqinn;             481;       668;      1224;
    zombie;            323;       557;      1050;


Summary
------------------------------------------------------------------------------

- We cannot declare a clear winner, it all depends on the use case.
- SQLite without CGO is possible nowadays.
