package app

import "time"

// Db is the database interface.
type Db interface {
	Exec(sqls ...string)
	InsertUsers(insertSql string, users []User)
	InsertArticles(insertSql string, articles []Article)
	InsertComments(insertSql string, comments []Comment)
	FindUsers(querySql string) []User
	FindUsersArticlesComments(querySql string) ([]User, []Article, []Comment)
	Close()
}

// User is a registered User who can access the blog.
type User struct {
	Id      int
	Created time.Time
	Email   string
	Active  bool
}

func NewUser(id int, created time.Time, email string, active bool) User {
	return User{id, created, email, active}
}

// Articles are created by Users.
type Article struct {
	Id      int
	Created time.Time
	UserId  int // the user that wrote this article
	Text    string
}

func NewArticle(id int, created time.Time, userId int, text string) Article {
	return Article{id, created, userId, text}
}

// Comments are written for Articles.
type Comment struct {
	Id        int
	Created   time.Time
	ArticleId int // the article that this comment is for
	Text      string
}

func NewComment(id int, created time.Time, articleId int, text string) Comment {
	return Comment{id, created, articleId, text}
}

func BindTime(v time.Time) int64 {
	if v.IsZero() {
		return 0
	}
	return v.UnixMilli()
}

func UnbindTime(v int64) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.UnixMilli(v)
}
