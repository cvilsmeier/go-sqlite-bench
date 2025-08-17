package main

import (
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/cvilsmeier/go-sqlite-bench/app"
)

func main() {
	app.Run(func(dbfile string) app.Db {
		return newDb(dbfile)
	})
}

type dbImpl struct {
	conn *sqlite3.Conn
}

var _ app.Db = (*dbImpl)(nil)

func newDb(dbfile string) app.Db {
	conn, err := sqlite3.Open(dbfile, sqlite3.OPEN_READWRITE|sqlite3.OPEN_CREATE|sqlite3.OPEN_NOMUTEX)
	app.MustBeNil(err)
	return &dbImpl{conn}
}

func (d *dbImpl) DriverName() string {
	return "bvinc"
}

func (d *dbImpl) Exec(sqls ...string) {
	for _, sql := range sqls {
		err := d.conn.Exec(sql)
		app.MustBeNil(err)
	}
}

func (d *dbImpl) Begin() {
	d.conn.Begin()
}

func (d *dbImpl) Commit() {
	d.conn.Commit()
}

func (d *dbImpl) InsertUsers(insertSql string, users []app.User) {
	stmt, err := d.conn.Prepare(insertSql)
	app.MustBeNil(err)
	for _, u := range users {
		err := stmt.Bind(u.Id, app.BindTime(u.Created), u.Email, u.Active)
		app.MustBeNil(err)
		_, err = stmt.Step()
		app.MustBeNil(err)
		err = stmt.Reset()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
}

func (d *dbImpl) InsertArticles(insertSql string, articles []app.Article) {
	stmt, err := d.conn.Prepare(insertSql)
	app.MustBeNil(err)
	for _, a := range articles {
		err := stmt.Bind(a.Id, app.BindTime(a.Created), a.UserId, a.Text)
		app.MustBeNil(err)
		_, err = stmt.Step()
		app.MustBeNil(err)
		err = stmt.Reset()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
}

func (d *dbImpl) InsertComments(insertSql string, comments []app.Comment) {
	stmt, err := d.conn.Prepare(insertSql)
	app.MustBeNil(err)
	for _, u := range comments {
		err := stmt.Bind(u.Id, app.BindTime(u.Created), u.ArticleId, u.Text)
		app.MustBeNil(err)
		_, err = stmt.Step()
		app.MustBeNil(err)
		err = stmt.Reset()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
}

func (d *dbImpl) FindUsers(querySql string) []app.User {
	stmt, err := d.conn.Prepare(querySql)
	app.MustBeNil(err)
	more, err := stmt.Step()
	app.MustBeNil(err)
	var users []app.User
	for more {
		id, ok, err := stmt.ColumnInt(0)
		app.MustBe(ok)
		app.MustBeNil(err)
		created, ok, err := stmt.ColumnInt64(1)
		app.MustBe(ok)
		app.MustBeNil(err)
		email, ok, err := stmt.ColumnText(2)
		app.MustBe(ok)
		app.MustBeNil(err)
		active, ok, err := stmt.ColumnInt(3)
		app.MustBe(ok)
		app.MustBeNil(err)
		user := app.NewUser(id, app.UnbindTime(created), email, active != 0)
		users = append(users, user)
		more, err = stmt.Step()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
	return users
}

func (d *dbImpl) FindArticles(querySql string) []app.Article {
	stmt, err := d.conn.Prepare(querySql)
	app.MustBeNil(err)
	more, err := stmt.Step()
	app.MustBeNil(err)
	var articles []app.Article
	for more {
		id, ok, err := stmt.ColumnInt(0)
		app.MustBe(ok)
		app.MustBeNil(err)
		created, ok, err := stmt.ColumnInt64(1)
		app.MustBe(ok)
		app.MustBeNil(err)
		userId, ok, err := stmt.ColumnInt(2)
		app.MustBe(ok)
		app.MustBeNil(err)
		text, ok, err := stmt.ColumnText(3)
		app.MustBe(ok)
		app.MustBeNil(err)
		article := app.NewArticle(id, app.UnbindTime(created), userId, text)
		articles = append(articles, article)
		more, err = stmt.Step()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
	return articles
}

func (d *dbImpl) FindUsersArticlesComments(querySql string, params []any) ([]app.User, []app.Article, []app.Comment) {
	// collections
	var users []app.User
	userIndexer := make(map[int]int)
	var articles []app.Article
	articleIndexer := make(map[int]int)
	var comments []app.Comment
	commentIndexer := make(map[int]int)
	// query
	stmt, err := d.conn.Prepare(querySql)
	app.MustBeNil(err)
	if len(params) > 0 {
		stmt.Bind(params...)
	}
	more, err := stmt.Step()
	app.MustBeNil(err)
	for more {
		{
			id, ok, err := stmt.ColumnInt(0)
			app.MustBe(ok)
			app.MustBeNil(err)
			created, ok, err := stmt.ColumnInt64(1)
			app.MustBe(ok)
			app.MustBeNil(err)
			email, ok, err := stmt.ColumnText(2)
			app.MustBe(ok)
			app.MustBeNil(err)
			active, ok, err := stmt.ColumnInt(3)
			app.MustBe(ok)
			app.MustBeNil(err)
			user := app.NewUser(id, app.UnbindTime(created), email, active != 0)
			_, found := userIndexer[user.Id]
			if !found {
				userIndexer[user.Id] = len(users)
				users = append(users, user)
			}
		}
		{
			id, ok, err := stmt.ColumnInt(4)
			app.MustBe(ok)
			app.MustBeNil(err)
			created, ok, err := stmt.ColumnInt64(5)
			app.MustBe(ok)
			app.MustBeNil(err)
			userId, ok, err := stmt.ColumnInt(6)
			app.MustBe(ok)
			app.MustBeNil(err)
			text, ok, err := stmt.ColumnText(7)
			app.MustBe(ok)
			app.MustBeNil(err)
			article := app.NewArticle(id, app.UnbindTime(created), userId, text)
			_, found := articleIndexer[article.Id]
			if !found {
				articleIndexer[article.Id] = len(articles)
				articles = append(articles, article)
			}
		}
		{
			id, ok, err := stmt.ColumnInt(8)
			app.MustBe(ok)
			app.MustBeNil(err)
			created, ok, err := stmt.ColumnInt64(9)
			app.MustBe(ok)
			app.MustBeNil(err)
			articleId, ok, err := stmt.ColumnInt(10)
			app.MustBe(ok)
			app.MustBeNil(err)
			text, ok, err := stmt.ColumnText(11)
			app.MustBe(ok)
			app.MustBeNil(err)
			comment := app.NewComment(id, app.UnbindTime(created), articleId, text)
			_, found := commentIndexer[comment.Id]
			if !found {
				commentIndexer[comment.Id] = len(comments)
				comments = append(comments, comment)
			}
		}
		more, err = stmt.Step()
		app.MustBeNil(err)
	}
	err = stmt.Close()
	app.MustBeNil(err)
	return users, articles, comments
}

func (d *dbImpl) Close() {
	err := d.conn.Close()
	app.MustBeNil(err)
}
