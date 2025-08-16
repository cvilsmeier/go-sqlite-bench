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


The test setup is as follows:

- OS: Debian/GNU Linux amd64 version 12.11
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 cores
- RAM: 32GB
- Disk: 1TB NVME SSD
- go version go1.24.5 linux/amd64

The benchmark was run on 2025-08-16, with current library versions,
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
    bvinc;         1121;    571;
    craw;          1249;    495;
    eaton;         1101;    614;
    glebarez;      5200;    763;
    mattn;         1572;   1051;
    modernc;       5102;    754;
    ncruces;       3064;    919;
    sqinn;          663;    237;
    zombie;        1764;    255;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    bvinc;          709;    682;
    craw;           732;    582;
    eaton;          723;    771;
    glebarez;      2899;   1088;
    mattn;          886;   1187;
    modernc;       2913;   1094;
    ncruces;       1854;   1220;
    sqinn;          457;    285;
    zombie;        1323;    488;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    bvinc;               22;          69;          545;
    craw;                14;          58;          508;
    eaton;               23;          73;          597;
    glebarez;            30;          91;          730;
    mattn;               29;         115;         1017;
    modernc;             30;          92;          751;
    ncruces;             39;         118;          984;
    sqinn;               42;          59;          334;
    zombie;              15;          33;          211;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    bvinc;                 171;            324;            521;
    craw;                  190;            304;            582;
    eaton;                 144;            240;            429;
    glebarez;              425;            696;           1104;
    mattn;                 133;            244;            469;
    modernc;               394;            661;           1141;
    ncruces;               181;            299;            562;
    sqinn;                 298;            543;           1115;
    zombie;                352;            560;            942;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    bvinc;             762;       967;      1768;
    craw;              576;       822;      1634;
    eaton;             715;      1018;      1940;
    glebarez;          938;      1163;      2146;
    mattn;            1164;      1502;      2835;
    modernc;           927;      1164;      2136;
    ncruces;          1077;      1338;      2572;
    sqinn;             425;       650;      1217;
    zombie;            351;       568;      1061;


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Zombiezen and Sqinn (both without CGO) are pretty fast.
