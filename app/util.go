package app

import (
	"fmt"
	"os"
	"time"
)

func Must(c bool, format string, a ...any) {
	if !c {
		panic(fmt.Sprintf("fail: "+format, a...))
	}
}

func MustBe(c bool) {
	if !c {
		panic("must be true but was false")
	}
}

func MustBeEqual(want, have any) {
	if want != have {
		panic(fmt.Sprintf("must be equal but was different:\nwant %#v\nhave %#v", want, have))
	}
}

func MustBeNil(a any) {
	if a != nil {
		panic(fmt.Sprintf("must be nil but was %#v", a))
	}
}

func MustBeSet(a any) {
	if a == nil {
		panic("must be set but was nil")
	}
}

func removeDbfiles(dbfile string) {
	// remove db file and temp files
	names := []string{dbfile, dbfile + "-shm", dbfile + "-wal", dbfile + "-journal"}
	for _, name := range names {
		os.Remove(name)
		_, err := os.Stat(name) // file must really be gone
		MustBeSet(err)
	}
}

func dbsize(dbfile string) int64 {
	var total int64
	names := []string{dbfile, dbfile + "-shm", dbfile + "-wal", dbfile + "-journal"}
	for _, name := range names {
		fi, err := os.Stat(name)
		if err == nil {
			total += fi.Size()
		}
	}
	return total
}

func millisSince(t time.Time) int64 {
	return time.Since(t).Milliseconds()
}
