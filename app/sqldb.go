package app

import (
	"database/sql"
)

// SqlDb is a Db implementation that uses database/sql package.
type SqlDb struct {
	db *sql.DB
}

var _ Db = (*SqlDb)(nil)

func NewSqlDb(db *sql.DB) *SqlDb {
	return &SqlDb{db}
}

func (d *SqlDb) Exec(sqls ...string) {
	for _, s := range sqls {
		_, err := d.db.Exec(s)
		MustBeNil(err)
	}
}

func (d *SqlDb) InsertUsers(insertSql string, users []User) {
	tx, err := d.db.Begin()
	MustBeNil(err)
	stmt, err := tx.Prepare(insertSql)
	MustBeNil(err)
	for _, u := range users {
		_, err = stmt.Exec(u.Id, BindTime(u.Created), u.Email, u.Active)
		MustBeNil(err)
	}
	err = stmt.Close()
	MustBeNil(err)
	err = tx.Commit()
	MustBeNil(err)
}

func (d *SqlDb) InsertArticles(insertSql string, articles []Article) {
	tx, err := d.db.Begin()
	MustBeNil(err)
	stmt, err := tx.Prepare(insertSql)
	MustBeNil(err)
	for _, u := range articles {
		_, err = stmt.Exec(u.Id, BindTime(u.Created), u.UserId, u.Text)
		MustBeNil(err)
	}
	err = stmt.Close()
	MustBeNil(err)
	err = tx.Commit()
	MustBeNil(err)
}

func (d *SqlDb) InsertComments(insertSql string, comments []Comment) {
	tx, err := d.db.Begin()
	MustBeNil(err)
	stmt, err := tx.Prepare(insertSql)
	MustBeNil(err)
	for _, u := range comments {
		_, err = stmt.Exec(u.Id, BindTime(u.Created), u.ArticleId, u.Text)
		MustBeNil(err)
	}
	err = stmt.Close()
	MustBeNil(err)
	err = tx.Commit()
	MustBeNil(err)
}

func (d *SqlDb) FindUsers(querySql string) []User {
	rows, err := d.db.Query(querySql)
	MustBeNil(err)
	var id sql.NullInt32
	var created sql.NullInt64
	var email sql.NullString
	var active sql.NullBool
	var users []User
	for rows.Next() {
		err = rows.Scan(&id, &created, &email, &active)
		MustBeNil(err)
		users = append(users, NewUser(int(id.Int32), UnbindTime(created.Int64), email.String, active.Bool))
	}
	return users
}

func (d *SqlDb) FindArticles(querySql string) []Article {
	rows, err := d.db.Query(querySql)
	MustBeNil(err)
	var id sql.NullInt32
	var created sql.NullInt64
	var userId sql.NullInt32
	var text sql.NullString
	var articles []Article
	for rows.Next() {
		err = rows.Scan(&id, &created, &userId, &text)
		MustBeNil(err)
		articles = append(articles, NewArticle(int(id.Int32), UnbindTime(created.Int64), int(userId.Int32), text.String))
	}
	return articles
}

func (d *SqlDb) FindUsersArticlesComments(querySql string) ([]User, []Article, []Comment) {
	rows, err := d.db.Query(querySql)
	MustBeNil(err)
	var userId sql.NullInt32
	var userCreated sql.NullInt64
	var userEmail sql.NullString
	var userActive sql.NullBool
	var articleId sql.NullInt32
	var articleCreated sql.NullInt64
	var articleUserId sql.NullInt32
	var articleText sql.NullString
	var commentId sql.NullInt32
	var commentCreated sql.NullInt64
	var commentArticleId sql.NullInt32
	var commentText sql.NullString
	// collections
	var users []User
	userIndexer := make(map[int]int)
	var articles []Article
	articleIndexer := make(map[int]int)
	var comments []Comment
	commentIndexer := make(map[int]int)
	for rows.Next() {
		err = rows.Scan(&userId, &userCreated, &userEmail, &userActive,
			&articleId, &articleCreated, &articleUserId, &articleText,
			&commentId, &commentCreated, &commentArticleId, &commentText)
		MustBeNil(err)
		user := NewUser(int(userId.Int32), UnbindTime(userCreated.Int64), userEmail.String, userActive.Bool)
		article := NewArticle(int(articleId.Int32), UnbindTime(articleCreated.Int64), int(articleUserId.Int32), articleText.String)
		comment := NewComment(int(commentId.Int32), UnbindTime(commentCreated.Int64), int(commentArticleId.Int32), commentText.String)
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

func (d *SqlDb) Close() {
	err := d.db.Close()
	MustBeNil(err)
}
