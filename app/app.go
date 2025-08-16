package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const verbose = false

func Run(makeDb func(dbfile string) Db) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.Print("")
	benchmarks := "simple,complex,many,large,concurrent"
	flag.StringVar(&benchmarks, "benchmarks", benchmarks, "specify benchmarks to run, comma separated")
	flag.Parse()
	dbfile := flag.Arg(0)
	if dbfile == "" {
		log.Fatal("dbfile empty, cannot bench")
	}
	// verbose
	if verbose {
		log.Printf("dbfile %q", dbfile)
	}
	// run benchmarks
	if strings.Contains(benchmarks, "simple") {
		benchSimple(dbfile, makeDb)
	}
	if strings.Contains(benchmarks, "complex") {
		benchComplex(dbfile, makeDb)
	}
	if strings.Contains(benchmarks, "many") {
		benchMany(dbfile, 10, makeDb)
		benchMany(dbfile, 100, makeDb)
		benchMany(dbfile, 1_000, makeDb)
	}
	if strings.Contains(benchmarks, "large") {
		benchLarge(dbfile, 50_000, makeDb)
		benchLarge(dbfile, 100_000, makeDb)
		benchLarge(dbfile, 200_000, makeDb)
	}
	if strings.Contains(benchmarks, "concurrent") {
		benchConcurrent(dbfile, 2, makeDb)
		benchConcurrent(dbfile, 4, makeDb)
		benchConcurrent(dbfile, 8, makeDb)
	}
}

const insertUserSql = "INSERT INTO users(id,created,email,active) VALUES(?,?,?,?)"
const insertArticleSql = "INSERT INTO articles(id,created,userId,text) VALUES(?,?,?,?)"
const insertCommentSql = "INSERT INTO comments(id,created,articleId,text) VALUES(?,?,?,?)"

func initSchema(db Db) {
	db.Exec(
		"PRAGMA journal_mode=DELETE",
		"PRAGMA synchronous=FULL",
		"PRAGMA foreign_keys=1",
		"PRAGMA busy_timeout=5000", // 5s busy timeout
		"CREATE TABLE users ("+
			"id INTEGER PRIMARY KEY NOT NULL,"+
			" created INTEGER NOT NULL,"+ // time.Time
			" email TEXT NOT NULL,"+
			" active INTEGER NOT NULL)", // bool
		"CREATE INDEX users_created ON users(created)",
		"CREATE TABLE articles ("+
			"id INTEGER PRIMARY KEY NOT NULL,"+
			" created INTEGER NOT NULL, "+ // time.Time
			" userId INTEGER NOT NULL REFERENCES users(id),"+
			" text TEXT NOT NULL)",
		"CREATE INDEX articles_created ON articles(created)",
		"CREATE INDEX articles_userId ON articles(userId)",
		"CREATE TABLE comments ("+
			"id INTEGER PRIMARY KEY NOT NULL,"+
			" created INTEGER NOT NULL, "+ // time.Time
			" articleId INTEGER NOT NULL REFERENCES articles(id),"+
			" text TEXT NOT NULL)",
		"CREATE INDEX comments_created ON comments(created)",
		"CREATE INDEX comments_articleId ON comments(articleId)",
	)
}

// Insert 1 million user rows in one database transaction.
// Then query all users once.
func benchSimple(dbfile string, makeDb func(dbfile string) Db) {
	removeDbfiles(dbfile)
	db := makeDb(dbfile)
	defer db.Close()
	initSchema(db)
	// insert users
	var users []User
	base := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	const nusers = 1_000_000
	for i := range nusers {
		users = append(users, NewUser(
			i+1,                                      // id,
			base.Add(time.Duration(i)*time.Minute),   // created,
			fmt.Sprintf("user%08d@example.com", i+1), // email,
			true,                                     // active,
		))
	}
	t0 := time.Now()
	db.InsertUsers("INSERT INTO users(id,created,email,active) VALUES(?,?,?,?)", users)
	insertMillis := millisSince(t0)
	if verbose {
		log.Printf("  insert took %d ms", insertMillis)
	}
	// query users
	t0 = time.Now()
	users = db.FindUsers("SELECT id,created,email,active FROM users ORDER BY id")
	MustBeEqual(len(users), nusers)
	queryMillis := millisSince(t0)
	if verbose {
		log.Printf("  query took %d ms", queryMillis)
	}
	// validate query result
	for i, u := range users {
		MustBeEqual(i+1, u.Id)
		MustBeEqual(base.Add(time.Duration(i)*time.Minute), u.Created)
		MustBeEqual(fmt.Sprintf("user%08d@example.com", i+1), u.Email)
		MustBeEqual(true, u.Active)
	}
	// print results
	bench := "1_simple"
	log.Printf("%s - insert - %-10s - %10d", bench, db.DriverName(), insertMillis)
	log.Printf("%s - query  - %-10s - %10d", bench, db.DriverName(), queryMillis)
	log.Printf("%s - dbsize - %-10s - %10d", bench, db.DriverName(), dbsize(dbfile))
}

// Insert 200 users in one database transaction.
// Then insert 20000 articles (100 articles for each user) in another transaction.
// Then insert 400000 articles (20 comments for each article) in another transaction.
// Then query all users, articles and comments in one big JOIN statement.
func benchComplex(dbfile string, makeDb func(dbfile string) Db) {
	removeDbfiles(dbfile)
	db := makeDb(dbfile)
	defer db.Close()
	initSchema(db)
	const nusers = 200
	const narticlesPerUser = 100
	const ncommentsPerArticle = 20
	if verbose {
		log.Printf("nusers = %d", nusers)
		log.Printf("narticlesPerUser = %d", narticlesPerUser)
		log.Printf("ncommentsPerArticle = %d", ncommentsPerArticle)
	}
	// make users, articles, comments
	var users []User
	var articles []Article
	var comments []Comment
	base := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	var userId int
	var articleId int
	var commentId int
	for range nusers {
		userId++
		user := NewUser(
			userId, // id
			base.Add(time.Duration(userId)*time.Minute), // created
			fmt.Sprintf("user%08d@example.com", userId), // email
			userId%2 == 0, // active
		)
		users = append(users, user)
		for range narticlesPerUser {
			articleId++
			article := NewArticle(
				articleId, // id
				base.Add(time.Duration(articleId)*time.Minute), // created
				userId,         // userId
				"article text", // text
			)
			articles = append(articles, article)
			for range ncommentsPerArticle {
				commentId++
				comment := NewComment(
					commentId, // id
					base.Add(time.Duration(commentId)*time.Minute), // created
					articleId,      // articleId
					"comment text", // text,
				)
				comments = append(comments, comment)
			}
		}
	}
	// insert users, articles, comments
	t0 := time.Now()
	db.InsertUsers(insertUserSql, users)
	db.InsertArticles(insertArticleSql, articles)
	db.InsertComments(insertCommentSql, comments)
	insertMillis := millisSince(t0)
	if verbose {
		log.Printf("  insert took %d ms", insertMillis)
	}
	// query users, articles, comments in one big join
	querySql := "SELECT" +
		" users.id, users.created, users.email, users.active," +
		" articles.id, articles.created, articles.userId, articles.text," +
		" comments.id, comments.created, comments.articleId, comments.text" +
		" FROM users" +
		" LEFT JOIN articles ON articles.userId = users.id" +
		" LEFT JOIN comments ON comments.articleId = articles.id" +
		" ORDER BY users.created,  articles.created, comments.created"
	t0 = time.Now()
	users, articles, comments = db.FindUsersArticlesComments(querySql)
	queryMillis := millisSince(t0)
	if verbose {
		log.Printf("  query took %d ms", queryMillis)
	}
	// validate query result
	MustBeEqual(nusers, len(users))
	MustBeEqual(nusers*narticlesPerUser, len(articles))
	MustBeEqual(nusers*narticlesPerUser*ncommentsPerArticle, len(comments))
	for i, user := range users {
		userId := i + 1
		MustBeEqual(userId, user.Id)
		MustBeEqual(base.Add(time.Duration(userId)*time.Minute), user.Created)
		MustBeEqual(fmt.Sprintf("user%08d@example.com", userId), user.Email)
		MustBeEqual(userId%2 == 0, user.Active)
	}
	for i, article := range articles {
		articleId := i + 1
		MustBeEqual(articleId, article.Id)
		MustBeEqual(base.Add(time.Duration(articleId)*time.Minute), article.Created)
		MustBe(1 <= article.UserId && article.UserId <= 1+nusers)
		MustBeEqual("article text", article.Text)
		if i > 0 {
			last := articles[i-1]
			MustBe(article.UserId >= last.UserId)
		}
	}
	for i, comment := range comments {
		commentId := i + 1
		MustBeEqual(commentId, comment.Id)
		MustBeEqual(base.Add(time.Duration(commentId)*time.Minute), comment.Created)
		MustBe(comment.ArticleId >= 1)
		MustBe(comment.ArticleId <= 1+(nusers*narticlesPerUser))
		MustBeEqual("comment text", comment.Text)
		if i > 0 {
			last := comments[i-1]
			MustBe(comment.ArticleId >= last.ArticleId)
		}
	}
	// print results
	bench := "2_complex"
	log.Printf("%s - insert - %-10s - %10d", bench, db.DriverName(), insertMillis)
	log.Printf("%s - query  - %-10s - %10d", bench, db.DriverName(), queryMillis)
	log.Printf("%s - dbsize - %-10s - %10d", bench, db.DriverName(), dbsize(dbfile))
}

// Insert N users in one database transaction.
// Then query all users 1000 times.
// This benchmark is used to simluate a read-heavy use case.
func benchMany(dbfile string, nusers int, makeDb func(dbfile string) Db) {
	removeDbfiles(dbfile)
	db := makeDb(dbfile)
	defer db.Close()
	initSchema(db)
	// insert users
	var users []User
	base := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	for iuser := range nusers {
		users = append(users, NewUser(
			iuser+1, // id,
			base.Add(time.Duration(iuser)*time.Minute),   // created,
			fmt.Sprintf("user%08d@example.com", iuser+1), // email,
			true, // active,
		))
	}
	t0 := time.Now()
	db.InsertUsers(insertUserSql, users)
	insertMillis := millisSince(t0)
	if verbose {
		log.Printf("  insert took %d ms", insertMillis)
	}
	// query users 1000 times
	t0 = time.Now()
	for i := 0; i < 1000; i++ {
		users = db.FindUsers("SELECT id,created,email,active FROM users ORDER BY id")
		MustBeEqual(len(users), nusers)
	}
	queryMillis := millisSince(t0)
	if verbose {
		log.Printf("  query took %d ms", queryMillis)
	}
	// validate query result
	for iuser, user := range users {
		MustBeEqual(iuser+1, user.Id)
		MustBeEqual(base.Add(time.Duration(iuser)*time.Minute), user.Created)
		MustBeEqual(fmt.Sprintf("user%08d@example.com", iuser+1), user.Email)
		MustBeEqual(true, user.Active)
	}
	// print results
	bench := fmt.Sprintf("3_many/%04d", nusers)
	log.Printf("%s - insert - %-10s - %10d", bench, db.DriverName(), insertMillis)
	log.Printf("%s - query  - %-10s - %10d", bench, db.DriverName(), queryMillis)
	log.Printf("%s - dbsize - %-10s - %10d", bench, db.DriverName(), dbsize(dbfile))
}

// Insert 10000 users with N bytes of row content.
// Then query all users.
// This benchmark is used to simluate reading of large (gigabytes) databases.
func benchLarge(dbfile string, nsize int, makeDb func(dbfile string) Db) {
	removeDbfiles(dbfile)
	db := makeDb(dbfile)
	defer db.Close()
	initSchema(db)
	// insert user with large emails
	t0 := time.Now()
	base := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	const nusers = 10_000
	var users []User
	for i := 0; i < nusers; i++ {
		users = append(users, NewUser(
			i+1,                                    // Id
			base.Add(time.Duration(i)*time.Second), // Created
			strings.Repeat("a", nsize),             // Email
			true,                                   // Active
		))
	}
	db.InsertUsers(insertUserSql, users)
	insertMillis := millisSince(t0)
	// query users
	t0 = time.Now()
	users = db.FindUsers("SELECT id,created,email,active FROM users ORDER BY id")
	MustBeEqual(len(users), nusers)
	queryMillis := millisSince(t0)
	if verbose {
		log.Printf("  query took %d ms", queryMillis)
	}
	// validate query result
	for i, u := range users {
		MustBeEqual(i+1, u.Id)
		MustBeEqual(2023, u.Created.Year())
		MustBeEqual("a", u.Email[0:1])
		MustBeEqual(true, u.Active)
	}
	// print results
	bench := fmt.Sprintf("4_large/%06d", nsize)
	log.Printf("%s - insert - %-10s - %10d", bench, db.DriverName(), insertMillis)
	log.Printf("%s - query  - %-10s - %10d", bench, db.DriverName(), queryMillis)
	log.Printf("%s - dbsize - %-10s - %10d", bench, db.DriverName(), dbsize(dbfile))
}

// Insert one million users.
// Then have N goroutines query all users.
// This benchmark is used to simulate concurrent reads.
func benchConcurrent(dbfile string, ngoroutines int, makeDb func(dbfile string) Db) {
	removeDbfiles(dbfile)
	db1 := makeDb(dbfile)
	driverName := db1.DriverName()
	initSchema(db1)
	// insert many users
	base := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	const nusers = 1_000_000
	var users []User
	for i := range nusers {
		users = append(users, NewUser(
			i+1,                                    // Id
			base.Add(time.Duration(i)*time.Second), // Created
			fmt.Sprintf("user%d@example.com", i+1), // Email
			true,                                   // Active
		))
	}
	t0 := time.Now()
	db1.InsertUsers(insertUserSql, users)
	db1.Close()
	insertMillis := millisSince(t0)
	// query users in N goroutines
	t0 = time.Now()
	var wg sync.WaitGroup
	for range ngoroutines {
		wg.Add(1)
		db := makeDb(dbfile)
		go func() {
			defer wg.Done()
			db.Exec(
				"PRAGMA foreign_keys=1",
				"PRAGMA busy_timeout=5000", // 5s busy timeout
			)
			defer db.Close()
			users = db.FindUsers("SELECT id,created,email,active FROM users ORDER BY id")
			MustBeEqual(len(users), nusers)
			// validate query result
			for i, u := range users {
				MustBeEqual(i+1, u.Id)
				MustBeEqual(2023, u.Created.Year())
				MustBeEqual("user", u.Email[0:4])
				MustBeEqual(true, u.Active)
			}
		}()
	}
	// wait for completion
	wg.Wait()
	queryMillis := millisSince(t0)
	if verbose {
		log.Printf("  query took %d ms", queryMillis)
	}
	// print results
	bench := fmt.Sprintf("5_concurrent/%d", ngoroutines)
	log.Printf("%s - insert - %-10s - %10d", bench, driverName, insertMillis)
	log.Printf("%s - query  - %-10s - %10d", bench, driverName, queryMillis)
	log.Printf("%s - dbsize - %-10s - %10d", bench, driverName, dbsize(dbfile))
}
