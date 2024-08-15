Benchmarks for Golang SQLite Drivers
==============================================================================

This work is sponsored by Monibot - Server and Application Monitoring.
Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).


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

- OS: Debian/GNU Linux amd64 version 12.6
- CPU: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz, 8 cores
- RAM: 16GB
- Disk: 1TB NVME SSD
- go version go1.22.6 linux/amd64

The benchmark was run on 2024-08-15, with current library versions,
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
    craw;          1211;    566;
    eaton;         1156;    713;
    mattn;         1641;   1119;
    modernc;       5662;   1170;
    ncruces;       3304;   1015;
    sqinn;          919;    590;
    zombie;        1932;    314;


### Complex

Insert 200 users in one database transaction.
Then insert 20000 articles (100 articles for each user) in another transaction.
Then insert 400000 comments (20 comments for each article) in another transaction.
Then query all users, articles and comments in one big JOIN statement.

![](results/complex.png)

    Complex;     insert;  query;
    craw;           701;    602;
    eaton;          733;    807;
    mattn;          891;   1268;
    modernc;       3112;   1515;
    ncruces;       1956;   1271;
    sqinn;          610;    787;
    zombie;        1470;    512;


### Many

Insert N users in one database transaction.
Then query all users 1000 times.
This benchmark is used to simluate a read-heavy use case.

![](results/many.png)

    Many;        query/N=10; query/N=100; query/N=1000;
    craw;                13;          61;          520;
    eaton;               25;          80;          637;
    mattn;               31;         109;          985;
    modernc;             35;         125;         1059;
    ncruces;             43;         117;          967;
    sqinn;               36;          69;          617;
    zombie;              17;          38;          215;


### Large

Insert 10000 users with N bytes of row content.
Then query all users.
This benchmark is used to simluate reading of large (gigabytes) databases.

![](results/large.png)

    Large;       query/N=50000; query/N=100000; query/N=200000;
    craw;                  198;            350;            672;
    eaton;                 182;            324;            583;
    mattn;                 158;            321;            516;
    modernc;               217;            698;           1245;
    ncruces;               190;            360;            648;
    sqinn;                 550;           1086;           2507;
    zombie;                169;            554;           1087;


### Concurrent

Insert one million users.
Then have N goroutines query all users.
This benchmark is used to simulate concurrent reads.

![](results/concurrent.png)

    Concurrent;  query/N=2; query/N=4; query/N=8;
    craw;              670;      1006;      1656;
    eaton;             830;      1299;      2036;
    mattn;            1287;      1606;      2917;
    modernc;          2811;      7046;     17849;
    ncruces;          1172;      1534;      2697;
    sqinn;             638;      1203;      2143;
    zombie;            342;       569;      1087;


Summary
------------------------------------------------------------------------------

- We cannot declare a winner, it all depends on the use case.
- Crawshaw and Zombiezen are pretty fast.
- SQLite without CGO is possible.


This work is sponsored by Monibot - Server and Application Monitoring.
Try out Monibot for free at [https://monibot.io](https://monibot.io?ref=go-sqlite-bench).
