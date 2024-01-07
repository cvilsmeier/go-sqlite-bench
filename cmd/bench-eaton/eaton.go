package main

import (
	"github.com/cvilsmeier/go-sqlite-bench/app"
	"github.com/eatonphil/gosqlite"
)

func main() {
	app.Run(func(dbfile string) app.Db {
		return newDb(dbfile)
	})
}

type dbImpl struct {
	conn *gosqlite.Conn
}

var _ app.Db = (*dbImpl)(nil)

func newDb(dbfile string) app.Db {
	flags := gosqlite.OPEN_READWRITE |
		gosqlite.OPEN_CREATE |
		gosqlite.OPEN_URI |
		gosqlite.OPEN_NOMUTEX
	conn, err := gosqlite.Open(dbfile, flags)
	app.MustBeNil(err)
	return &dbImpl{conn}
}

func (d *dbImpl) DriverName() string {
	return "eaton"
}

func (d *dbImpl) Exec(sqls ...string) {
	for _, s := range sqls {
		d.exec(s)
	}
}

func (d *dbImpl) exec(sql string) {
	err := d.conn.Exec(sql)
	app.MustBeNil(err)
}

func (d *dbImpl) prepare(sql string) *gosqlite.Stmt {
	stmt, err := d.conn.Prepare(sql)
	app.MustBeNil(err)
	return stmt
}

func (d *dbImpl) InsertUsers(insertSql string, users []app.User) {
	app.MustBeNil(d.conn.Begin())
	stmt := d.prepare(insertSql)
	for _, u := range users {
		err := stmt.Exec(int64(u.Id), app.BindTime(u.Created), u.Email, u.Active)
		app.MustBeNil(err)
	}
	app.MustBeNil(stmt.Close())
	app.MustBeNil(d.conn.Commit())
}

func (d *dbImpl) InsertArticles(insertSql string, articles []app.Article) {
	app.MustBeNil(d.conn.Begin())
	stmt := d.prepare(insertSql)
	for _, u := range articles {
		err := stmt.Exec(int64(u.Id), app.BindTime(u.Created), int64(u.UserId), u.Text)
		app.MustBeNil(err)
	}
	app.MustBeNil(stmt.Close())
	app.MustBeNil(d.conn.Commit())
}

func (d *dbImpl) InsertComments(insertSql string, comments []app.Comment) {
	app.MustBeNil(d.conn.Begin())
	stmt := d.prepare(insertSql)
	for _, u := range comments {
		err := stmt.Exec(int64(u.Id), app.BindTime(u.Created), int64(u.ArticleId), u.Text)
		app.MustBeNil(err)
	}
	app.MustBeNil(stmt.Close())
	app.MustBeNil(d.conn.Commit())
}

func (d *dbImpl) FindUsers(querySql string) []app.User {
	app.MustBeNil(d.conn.Begin())
	stmt := d.prepare(querySql)
	var users []app.User
	for {
		hasRow, err := stmt.Step()
		app.MustBeNil(err)
		if !hasRow {
			break
		}
		var user app.User
		var createdInt int64
		err = stmt.Scan(&user.Id, &createdInt, &user.Email, &user.Active)
		app.MustBeNil(err)
		user.Created = app.UnbindTime(createdInt)
		users = append(users, user)
	}
	app.MustBeNil(stmt.Close())
	app.MustBeNil(d.conn.Commit())
	return users
}

func (d *dbImpl) FindArticles(querySql string) []app.Article {
	app.MustBeNil(d.conn.Begin())
	stmt := d.prepare(querySql)
	var articles []app.Article
	for {
		hasRow, err := stmt.Step()
		app.MustBeNil(err)
		if !hasRow {
			break
		}
		var article app.Article
		var createdInt int64
		err = stmt.Scan(&article.Id, &createdInt, &article.UserId, &article.Text)
		app.MustBeNil(err)
		article.Created = app.UnbindTime(createdInt)
		articles = append(articles, article)
	}
	app.MustBeNil(stmt.Close())
	app.MustBeNil(d.conn.Commit())
	return articles
}

func (d *dbImpl) FindUsersArticlesComments(querySql string) ([]app.User, []app.Article, []app.Comment) {
	stmt := d.prepare(querySql)
	// collections
	var users []app.User
	userIndexer := make(map[int]int)
	var articles []app.Article
	articleIndexer := make(map[int]int)
	var comments []app.Comment
	commentIndexer := make(map[int]int)
	for {
		hasRow, err := stmt.Step()
		app.MustBeNil(err)
		if !hasRow {
			break
		}
		var user app.User
		var article app.Article
		var comment app.Comment
		var userCreated, articleCreated, commentCreated int64
		err = stmt.Scan(
			&user.Id, &userCreated, &user.Email, &user.Active,
			&article.Id, &articleCreated, &article.UserId, &article.Text,
			&comment.Id, &commentCreated, &comment.ArticleId, &comment.Text,
		)
		app.MustBeNil(err)
		user.Created = app.UnbindTime(userCreated)
		article.Created = app.UnbindTime(articleCreated)
		comment.Created = app.UnbindTime(commentCreated)
		_, ok := userIndexer[user.Id]
		if !ok {
			userIndexer[user.Id] = len(users)
			users = append(users, user)
		}
		_, ok = articleIndexer[article.Id]
		if !ok {
			articleIndexer[article.Id] = len(articles)
			articles = append(articles, article)
		}
		_, ok = commentIndexer[comment.Id]
		if !ok {
			commentIndexer[comment.Id] = len(comments)
			comments = append(comments, comment)
		}
	}
	app.MustBeNil(stmt.Close())
	return users, articles, comments
}

func (d *dbImpl) Close() {
	err := d.conn.Close()
	app.MustBeNil(err)
}
