module github.com/cvilsmeier/sqinn-go-bench

go 1.14

// cvvvvvvvv remove replace
replace github.com/cvilsmeier/sqinn-go => C:/Eigene/projects/vilsmeier/sqinn/dev/sqinn-go

require (
	crawshaw.io/sqlite v0.3.2
	github.com/cvilsmeier/sqinn-go v1.1.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
)
