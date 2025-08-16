package main

import (
	"github.com/cvilsmeier/go-sqlite-bench/app"
	"github.com/cvilsmeier/sqinn-go/v2"
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
	sq := sqinn.MustLaunch(sqinn.Options{Db: dbfile})
	sq.MustExecSql("PRAGMA foreign_keys=1")
	return &dbImpl{sq}
}

func (d *dbImpl) DriverName() string {
	return "sqinn"
}

func (d *dbImpl) Exec(sqls ...string) {
	for _, s := range sqls {
		d.sq.MustExecSql(s)
	}
}

func (d *dbImpl) InsertUsers(insertSql string, users []app.User) {
	d.sq.MustExecSql("BEGIN")
	err := d.sq.Exec(insertSql, len(users), 4, func(iteration int, params []sqinn.Value) {
		user := users[iteration]
		params[0].Type = sqinn.ValInt32
		params[0].Int32 = user.Id
		params[1].Type = sqinn.ValInt64
		params[1].Int64 = app.BindTime(user.Created)
		params[2].Type = sqinn.ValString
		params[2].String = user.Email
		params[3].Type = sqinn.ValInt32
		params[3].Int32 = bindBool(user.Active)
	})
	if err != nil {
		panic(err)
	}
	d.sq.MustExecSql("COMMIT")
}

func (d *dbImpl) InsertArticles(insertSql string, articles []app.Article) {
	d.sq.MustExecSql("BEGIN")
	err := d.sq.Exec(insertSql, len(articles), 4, func(iteration int, params []sqinn.Value) {
		article := articles[iteration]
		params[0].Type = sqinn.ValInt32
		params[0].Int32 = article.Id
		params[1].Type = sqinn.ValInt64
		params[1].Int64 = app.BindTime(article.Created)
		params[2].Type = sqinn.ValInt32
		params[2].Int32 = article.UserId
		params[3].Type = sqinn.ValString
		params[3].String = article.Text
	})
	if err != nil {
		panic(err)
	}
	d.sq.MustExecSql("COMMIT")
}

func (d *dbImpl) InsertComments(insertSql string, comments []app.Comment) {
	d.sq.MustExecSql("BEGIN")
	err := d.sq.Exec(insertSql, len(comments), 4, func(iteration int, params []sqinn.Value) {
		comment := comments[iteration]
		params[0].Type = sqinn.ValInt32
		params[0].Int32 = comment.Id
		params[1].Type = sqinn.ValInt64
		params[1].Int64 = app.BindTime(comment.Created)
		params[2].Type = sqinn.ValInt32
		params[2].Int32 = comment.ArticleId
		params[3].Type = sqinn.ValString
		params[3].String = comment.Text
	})
	if err != nil {
		panic(err)
	}
	d.sq.MustExecSql("COMMIT")
}

func (d *dbImpl) FindUsers(querySql string) []app.User {
	users := make([]app.User, 0, 2*1024)
	coltypes := []byte{
		sqinn.ValInt32, sqinn.ValInt64, sqinn.ValString, sqinn.ValInt32, // User
	}
	err := d.sq.Query(querySql, nil, coltypes, func(row int, values []sqinn.Value) {
		users = append(users, readUser(values, 0))
	})
	if err != nil {
		panic(err)
	}
	return users
}

func (d *dbImpl) FindUsersArticlesComments(querySql string) ([]app.User, []app.Article, []app.Comment) {
	users := make([]app.User, 0, 2*1024)
	articles := make([]app.Article, 0, 2*1024)
	comments := make([]app.Comment, 0, 2*1024)
	coltypes := []byte{
		sqinn.ValInt32, sqinn.ValInt64, sqinn.ValString, sqinn.ValInt32, // User
		sqinn.ValInt32, sqinn.ValInt64, sqinn.ValInt32, sqinn.ValString, // Article
		sqinn.ValInt32, sqinn.ValInt64, sqinn.ValInt32, sqinn.ValString, // Comment
	}
	userIds := map[int]struct{}{}
	articleIds := map[int]struct{}{}
	commentIds := map[int]struct{}{}
	err := d.sq.Query(querySql, nil, coltypes, func(row int, values []sqinn.Value) {
		user := readUser(values, 0)
		article := readArticle(values, 4)
		comment := readComment(values, 8)
		if _, ok := userIds[user.Id]; !ok {
			userIds[user.Id] = struct{}{}
			users = append(users, user)
		}
		if _, ok := articleIds[article.Id]; !ok {
			articleIds[article.Id] = struct{}{}
			articles = append(articles, article)
		}
		if _, ok := commentIds[comment.Id]; !ok {
			commentIds[comment.Id] = struct{}{}
			comments = append(comments, comment)
		}
	})
	if err != nil {
		panic(err)
	}
	return users, articles, comments
}

func (d *dbImpl) Close() {
	err := d.sq.Close()
	app.MustBeNil(err)
}

func readUser(values []sqinn.Value, icol int) app.User {
	return app.NewUser(
		values[icol+0].Int32,                 // id int,
		app.UnbindTime(values[icol+1].Int64), // created time.Time,
		values[icol+2].String,                // email string,
		unbindBool(values[icol+3].Int32),     // active bool,
	)
}

func readArticle(values []sqinn.Value, icol int) app.Article {
	return app.NewArticle(
		values[icol+0].Int32,                 // id int,
		app.UnbindTime(values[icol+1].Int64), // created time.Time,
		values[icol+2].Int32,                 // userId int,
		values[icol+3].String,                // text string,
	)
}

func readComment(values []sqinn.Value, off int) app.Comment {
	return app.NewComment(
		values[off+0].Int32,                 // id int,
		app.UnbindTime(values[off+1].Int64), // created time.Time,
		values[off+2].Int32,                 // articleId int,
		values[off+3].String,                // text string,
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
