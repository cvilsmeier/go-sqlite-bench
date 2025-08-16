Benchmarks for Golang SQLite Drivers
==============================================================================

> [!NOTE]
> This work is sponsored by Monibot - Website, Server and Application Monitoring.
> Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).


For benchmarks I used the following libraries:

- craw, [github.com/crawshaw/sqlite](https://github.com/crawshaw/sqlite), a CGO-based solution. This is not a `database/sql` driver.

- eaton, [github.com/eatonphil/gosqlite](https://github.com/eatonphil/gosqlite), a CGO-based solution. This is not a
    `database/sql` driver. (addded by @c4rlo)

- glebarez, [github.com/glebarez/go-sqlite](https://github.com/glebarez/go-sqlite), a pure Go solution. This is a newer library,
    based on the SQLite C code re-written in Go (added by @dcarbone).

- mattn, [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3), a CGO-based solution. This library is
    (still) the de-facto standard and widely used. 

- modernc, [modernc.org/sqlite](https://modernc.org/sqlite), a pure Go solution. This is a newer library,
    based on the SQLite C code re-written in Go.

- ncruces, [github.com/ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3), a pure Go solution based on WASM (?). 

- sqinn, [github.com/cvilsmeier/sqinn-go](https://github.com/cvilsmeier/sqinn-go/tree/v1) (v1), a solution without CGO. It uses
    [github.com/cvilsmeier/sqinn](https://github.com/cvilsmeier/sqinn/tree/v1) (v1) to access SQLite database files.

- sqinn2, [github.com/cvilsmeier/sqinn-go](https://github.com/cvilsmeier/sqinn-go) (v2), a solution without CGO. It uses
    [github.com/cvilsmeier/sqinn](https://github.com/cvilsmeier/sqinn) (v2) to access SQLite database files.

- zombie, [github.com/zombiezen/go-sqlite](https://github.com/zombiezen/go-sqlite), a rewrite of the crawshaw driver, using the
    modernc libraries. This is not a `database/sql` driver.


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
    craw;          1219;    490;
    eaton;         1177;    603;
    glebarez;      5330;    759;
    mattn;         1582;   1051;
    modernc;       5383;    765;
    ncruces;       3087;    907;
    sqinn;          875;    645;
    sqinn2;         711;    230;
    zombie;        1761;    261;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    craw;           705;    608;
    eaton;          700;    747;
    glebarez;      2894;   1088;
    mattn;          881;   1208;
    modernc;       2926;   1113;
    ncruces;       1838;   1211;
    sqinn;          562;    725;
    sqinn2;         476;    289;
    zombie;        1276;    485;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    craw;                14;          58;          508;
    eaton;               24;          74;          589;
    glebarez;            29;          93;          747;
    mattn;               29;         120;         1009;
    modernc;             29;          91;          717;
    ncruces;             33;         113;          965;
    sqinn;               19;          66;          628;
    sqinn2;              40;          64;          342;
    zombie;              16;          35;          245;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    craw;                  185;            323;            584;
    eaton;                 144;            241;            424;
    glebarez;              408;            676;           1104;
    mattn;                 127;            263;            426;
    modernc;               415;            677;           1085;
    ncruces;               193;            303;            565;
    sqinn;                 554;           1069;           1979;
    sqinn2;                293;            539;           1132;
    zombie;                335;            569;            934;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    craw;              579;       850;      1598;
    eaton;             734;      1005;      1922;
    glebarez;          877;      1166;      2145;
    mattn;            1185;      1492;      2829;
    modernc;           846;      1180;      2146;
    ncruces;          1045;      1303;      2517;
    sqinn;             619;      1124;      2332;
    sqinn2;            426;       660;      1231;
    zombie;            341;       577;      1027;


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Zombiezen and Sqinn v2 (both without cgo) are pretty fast.
