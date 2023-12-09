package main

import (
	"database/sql"

	"github.com/cvilsmeier/go-sqlite-bench/app"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
	app.Run(func(dbfile string) app.Db {
		db, err := sql.Open("sqlite3", dbfile)
		app.MustBeNil(err)
		return app.NewSqlDb(db)
	})
}
