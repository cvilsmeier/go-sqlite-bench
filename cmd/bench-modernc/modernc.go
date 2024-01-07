package main

import (
	"database/sql"

	"github.com/cvilsmeier/go-sqlite-bench/app"
	_ "modernc.org/sqlite"
)

func main() {
	app.Run(func(dbfile string) app.Db {
		db, err := sql.Open("sqlite", dbfile)
		app.MustBeNil(err)
		return app.NewSqlDb("modernc", db)
	})
}
