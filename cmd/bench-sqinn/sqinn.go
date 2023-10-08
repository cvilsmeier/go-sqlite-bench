package main

import (
	"os"

	"github.com/cvilsmeier/go-sqlite-bench/app"
	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func main() {
	app.Run(func(dbfile string) app.Db {
		return newDb(dbfile)
	})
}

type dbImpl struct {
	sq *sqinn.Sqinn
}

var _ app.Db = (*dbImpl)(nil)

func newDb(dbfile string) app.Db {
	sq := sqinn.MustLaunch(sqinn.Options{SqinnPath: os.Getenv("SQINN_PATH")})
	sq.MustOpen(dbfile)
	sq.MustExecOne("PRAGMA foreign_keys=1")
	return &dbImpl{sq}
}

func (d *dbImpl) Exec(sqls ...string) {
	for _, s := range sqls {
		d.sq.MustExecOne(s)
	}
}

func (d *dbImpl) InsertUsers(insertSql string, users []app.User) {
	d.sq.MustExecOne("BEGIN")
	const nparams = 4
	values := make([]any, 0, nparams*len(users))
	for _, u := range users {
		values = append(values,
			u.Id,
			app.BindTime(u.Created),
			u.Email,
			bindBool(u.Active),
		)
	}
	d.sq.MustExec(insertSql, len(users), nparams, values)
	d.sq.MustExecOne("COMMIT")
}

func (d *dbImpl) InsertArticles(insertSql string, articles []app.Article) {
	d.sq.MustExecOne("BEGIN")
	const nparams = 4
	values := make([]any, 0, nparams*len(articles))
	for _, u := range articles {
		values = append(values,
			u.Id,
			app.BindTime(u.Created),
			u.UserId,
			u.Text,
		)
	}
	d.sq.MustExec(insertSql, len(articles), nparams, values)
	d.sq.MustExecOne("COMMIT")
}

func (d *dbImpl) InsertComments(insertSql string, comments []app.Comment) {
	d.sq.MustExecOne("BEGIN")
	const nparams = 4
	values := make([]any, 0, nparams*len(comments))
	for _, u := range comments {
		values = append(values,
			u.Id,
			app.BindTime(u.Created),
			u.ArticleId,
			u.Text,
		)
	}
	d.sq.MustExec(insertSql, len(comments), nparams, values)
	d.sq.MustExecOne("COMMIT")
}

func (d *dbImpl) FindUsers(querySql string) []app.User {
	rows := d.sq.MustQuery(querySql, nil, []byte{sqinn.ValInt, sqinn.ValInt64, sqinn.ValText, sqinn.ValInt, sqinn.ValInt64})
	users := make([]app.User, len(rows))
	for i, row := range rows {
		users[i] = readUser(row.Values, 0)
	}
	return users
}

func (d *dbImpl) FindUsersArticlesComments(querySql string) ([]app.User, []app.Article, []app.Comment) {
	coltypes := []byte{
		sqinn.ValInt, sqinn.ValInt64, sqinn.ValText, sqinn.ValInt, // User
		sqinn.ValInt, sqinn.ValInt64, sqinn.ValInt, sqinn.ValText, // Article
		sqinn.ValInt, sqinn.ValInt64, sqinn.ValInt, sqinn.ValText, // Comment
	}
	rows := d.sq.MustQuery(querySql, nil, coltypes)
	// collections
	var users []app.User
	userIndexer := make(map[int]int)
	var articles []app.Article
	articleIndexer := make(map[int]int)
	var comments []app.Comment
	commentIndexer := make(map[int]int)
	for _, row := range rows {
		user := readUser(row.Values, 0)
		article := readArticle(row.Values, 4)
		comment := readComment(row.Values, 8)
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
	return users, articles, comments
}

func (d *dbImpl) Close() {
	err := d.sq.Close()
	app.MustBeNil(err)
	err = d.sq.Terminate()
	app.MustBeNil(err)
}

func readUser(values []sqinn.AnyValue, off int) app.User {
	return app.NewUser(
		values[off+0].AsInt(),                   // id int,
		app.UnbindTime(values[off+1].AsInt64()), // created time.Time,
		values[off+2].AsString(),                // email string,
		unbindBool(values[off+3].AsInt()),       // active bool,
	)
}

func readArticle(values []sqinn.AnyValue, off int) app.Article {
	return app.NewArticle(
		values[off+0].AsInt(),                   // id int,
		app.UnbindTime(values[off+1].AsInt64()), // created time.Time,
		values[off+2].AsInt(),                   // userId int,
		values[off+3].AsString(),                // text string,
	)
}

func readComment(values []sqinn.AnyValue, off int) app.Comment {
	return app.NewComment(
		values[off+0].AsInt(),                   // id int,
		app.UnbindTime(values[off+1].AsInt64()), // created time.Time,
		values[off+2].AsInt(),                   // articleId int,
		values[off+3].AsString(),                // text string,
	)
}

func bindBool(b bool) int {
	if b {
		return 1
	}
	return 0
}

func unbindBool(v int) bool {
	return v != 0
}
