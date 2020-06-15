package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/cvilsmeier/sqinn-go-bench/utl"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func prepStepFin(con *sqlite.Conn, sql string) {
	stmt := con.Prep(sql)
	_, err := stmt.Step()
	check(err)
	err = stmt.Finalize()
	check(err)
}

func openPrepare(dbfile string, clearDbfile bool, poolSize int, sqls []string) *sqlitex.Pool {
	// make sure db doesn't exist
	if clearDbfile {
		utl.Remove(dbfile)
	}
	// open db
	flags := sqlite.SQLITE_OPEN_READWRITE |
		sqlite.SQLITE_OPEN_CREATE |
		sqlite.SQLITE_OPEN_URI |
		sqlite.SQLITE_OPEN_NOMUTEX
	pool, err := sqlitex.Open(dbfile, flags, poolSize)
	check(err)
	// prepare schema
	conn := pool.Get(nil)
	for _, s := range sqls {
		prepStepFin(conn, s)
	}
	pool.Put(conn)
	return pool
}

func closePool(pool *sqlitex.Pool) {
	err := pool.Close()
	check(err)
}

func benchSimple(dbfile string, nusers int) {
	log.Printf("benchSimple dbfile=%s, nusers=%d", dbfile, nusers)
	pool := openPrepare(dbfile, true, 1, []string{utl.SqlCreateUsers})
	defer closePool(pool)
	conn := pool.Get(nil)
	defer pool.Put(conn)
	// insert users
	tstart := time.Now()
	prepStepFin(conn, "BEGIN")
	stmt := conn.Prep(utl.SqlInsertUsers)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User %d", id)
		age := 33 + id
		rating := 0.13 * float64(id)
		stmt.BindInt64(1, int64(id))
		stmt.BindText(2, name)
		stmt.BindInt64(3, int64(age))
		stmt.BindFloat(4, rating)
		_, err := stmt.Step()
		check(err)
		err = stmt.Reset()
		check(err)
	}
	err := stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	log.Printf("  insert took %s", utl.Since(tstart))
	// query users
	tstart = time.Now()
	stmt, err = conn.Prepare(utl.SqlSelectUsers)
	check(err)
	more, err := stmt.Step()
	check(err)
	var nrows int
	for more {
		nrows++
		var id int
		var name string
		var age int
		var rating float64
		if stmt.ColumnType(0) != sqlite.SQLITE_NULL {
			id = stmt.ColumnInt(0)
		}
		if stmt.ColumnType(1) != sqlite.SQLITE_NULL {
			name = stmt.ColumnText(1)
		}
		if stmt.ColumnType(2) != sqlite.SQLITE_NULL {
			age = stmt.ColumnInt(2)
		}
		if stmt.ColumnType(3) != sqlite.SQLITE_NULL {
			rating = stmt.ColumnFloat(3)
		}
		if id < 1 || len(name) < 5 || age < 33 || rating < 0.13 {
			log.Fatal("wrong row values")
		}
		more, err = stmt.Step()
		check(err)
	}
	if nrows != nusers {
		log.Fatalf("expected %v rows but was %v", nusers, nrows)
	}
	log.Printf("  query took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchComplex(dbfile string, nprofiles, ndevices, nlocations int) {
	log.Printf("benchComplex dbfile=%s, nprofiles, ndevices, nlocations = %d, %d, %d", dbfile, nprofiles, ndevices, nlocations)
	pool := openPrepare(dbfile, true, 1, utl.SqlCreateComplex)
	defer closePool(pool)
	conn := pool.Get(nil)
	defer pool.Put(conn)
	// insert profiles
	tstart := time.Now()
	prepStepFin(conn, "BEGIN")
	stmt := conn.Prep(utl.SqlInsertProfiles)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("Profile %d", p)
		active := p % 2
		stmt.BindText(1, profileID)
		stmt.BindText(2, name)
		stmt.BindInt64(3, int64(active))
		_, err := stmt.Step()
		check(err)
		err = stmt.Reset()
		check(err)
	}
	err := stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	// insert devices
	prepStepFin(conn, "BEGIN")
	stmt = conn.Prep(utl.SqlInsertDevices)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			name := fmt.Sprintf("Device %d %d", p, d)
			active := d % 2
			stmt.BindText(1, deviceID)
			stmt.BindText(2, profileID)
			stmt.BindText(3, name)
			stmt.BindInt64(4, int64(active))
			_, err = stmt.Step()
			check(err)
			err = stmt.Reset()
			check(err)
		}
	}
	err = stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	// insert locations
	prepStepFin(conn, "BEGIN")
	stmt = conn.Prep(utl.SqlInsertLocations)
	for p := 0; p < nprofiles; p++ {
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			for l := 0; l < nlocations; l++ {
				locationID := fmt.Sprintf("location_%d_%d_%d", p, d, l)
				name := fmt.Sprintf("Location %d %d %d", p, d, l)
				active := l % 2
				stmt.BindText(1, locationID)
				stmt.BindText(2, deviceID)
				stmt.BindText(3, name)
				stmt.BindInt64(4, int64(active))
				_, err = stmt.Step()
				check(err)
				err = stmt.Reset()
				check(err)
			}
		}
	}
	err = stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	log.Printf("  insert took %s", utl.Since(tstart))
	// query
	tstart = time.Now()
	stmt = conn.Prep(utl.SqlSelectComplex)
	stmt.BindInt64(1, int64(0))
	stmt.BindInt64(2, int64(1))
	more, err := stmt.Step()
	check(err)
	var nrows int
	for more {
		nrows++
		var locationID string
		var locationDeviceID string
		var locationName string
		var locationActive bool
		var deviceID string
		var deviceProfileID string
		var deviceName string
		var deviceActive bool
		var profileID string
		var profileName string
		var profileActive bool
		ci := 1
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			locationID = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			locationDeviceID = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			locationName = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			locationActive = stmt.ColumnInt(ci) != 0
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			deviceID = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			deviceProfileID = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			deviceName = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			deviceActive = stmt.ColumnInt(ci) != 0
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			profileID = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			profileName = stmt.ColumnText(ci)
		}
		ci++
		if stmt.ColumnType(ci) != sqlite.SQLITE_NULL {
			profileActive = stmt.ColumnInt(ci) != 0
		}
		_, _, _, _, _, _, _, _, _, _, _ = locationID, locationDeviceID, locationName, locationActive, deviceID, deviceProfileID, deviceName, deviceActive, profileID, profileName, profileActive
		more, err = stmt.Step()
		check(err)
	}
	expectedRows := nprofiles * ndevices * nlocations
	if nrows != expectedRows {
		log.Fatalf("expected %v rows but was %v", expectedRows, nrows)
	}
	log.Printf("  query took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchMany(dbfile string, ncars, nqueries int) {
	log.Printf("benchMany dbfile=%s, ncars=%d, nqueries=%d", dbfile, ncars, nqueries)
	pool := openPrepare(dbfile, true, 1, []string{utl.SqlCreateCars})
	defer closePool(pool)
	conn := pool.Get(nil)
	defer pool.Put(conn)
	// insert
	prepStepFin(conn, "BEGIN")
	stmt := conn.Prep(utl.SqlInsertCars)
	for i := 0; i < ncars; i++ {
		id := i + 1
		company := fmt.Sprintf("Company %d", id)
		model := fmt.Sprintf("Model %d", id)
		stmt.BindInt64(1, int64(id))
		stmt.BindText(2, company)
		stmt.BindText(3, model)
		_, err := stmt.Step()
		check(err)
		err = stmt.Reset()
		check(err)
	}
	err := stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		stmt, err = conn.Prepare(utl.SqlSelectCars)
		check(err)
		more, err := stmt.Step()
		check(err)
		var nrows int
		for more {
			nrows++
			var id int
			var company string
			var model string
			if stmt.ColumnType(0) != sqlite.SQLITE_NULL {
				id = stmt.ColumnInt(0)
			}
			if stmt.ColumnType(1) != sqlite.SQLITE_NULL {
				company = stmt.ColumnText(1)
			}
			if stmt.ColumnType(2) != sqlite.SQLITE_NULL {
				model = stmt.ColumnText(2)
			}
			if id < 1 || len(company) < 5 || len(model) < 5 {
				log.Fatal("wrong row values")
			}
			more, err = stmt.Step()
			check(err)
		}
		if nrows != ncars {
			log.Fatalf("expected %v rows but was %v", ncars, nrows)
		}
	}
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchLarge(dbfile string, nplants, nqueries, nameLength int) {
	log.Printf("benchLarge dbfile=%s, nplants=%d, nqueries=%d, nameLength=%d", dbfile, nplants, nqueries, nameLength)
	pool := openPrepare(dbfile, true, 1, []string{utl.SqlCreateCars})
	defer closePool(pool)
	conn := pool.Get(nil)
	defer pool.Put(conn)
	// insert
	name := ""
	for len(name) < nameLength {
		name = name + "Name "
	}
	prepStepFin(conn, "BEGIN")
	stmt := conn.Prep(utl.SqlInsertCars)
	for i := 0; i < nplants; i++ {
		id := i + 1
		stmt.BindInt64(1, int64(id))
		stmt.BindText(2, name)
		_, err := stmt.Step()
		check(err)
		err = stmt.Reset()
		check(err)
	}
	err := stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		stmt, err = conn.Prepare(utl.SqlSelectCars)
		check(err)
		more, err := stmt.Step()
		check(err)
		var nrows int
		for more {
			nrows++
			var id int
			var nameValue string
			if stmt.ColumnType(0) != sqlite.SQLITE_NULL {
				id = stmt.ColumnInt(0)
			}
			if stmt.ColumnType(1) != sqlite.SQLITE_NULL {
				nameValue = stmt.ColumnText(1)
			}
			if id < 1 || len(nameValue) < nameLength {
				log.Fatal("wrong row values")
			}
			more, err = stmt.Step()
			check(err)
		}
		if nrows != nplants {
			log.Fatalf("expected %v rows but was %v", nplants, nrows)
		}
	}
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchConcurrent(dbfile string, nbooks, nworkers int) {
	log.Printf("benchConcurrent dbfile=%s, nbooks=%d, nworkers=%d", dbfile, nbooks, nworkers)
	pool := openPrepare(dbfile, true, nworkers, []string{utl.SqlCreateBooks})
	defer closePool(pool)
	// insert
	tstart := time.Now()
	conn := pool.Get(nil)
	prepStepFin(conn, "BEGIN")
	stmt := conn.Prep(utl.SqlInsertBooks)
	for i := 0; i < nbooks; i++ {
		id := i + 1
		name := fmt.Sprintf("Book %d", id)
		stmt.BindInt64(1, int64(id))
		stmt.BindText(2, name)
		_, err := stmt.Step()
		check(err)
		err = stmt.Reset()
		check(err)
	}
	err := stmt.Finalize()
	check(err)
	prepStepFin(conn, "COMMIT")
	pool.Put(conn)
	// queries
	tstart = time.Now()
	var wg sync.WaitGroup
	wg.Add(nworkers)
	for w := 0; w < nworkers; w++ {
		go func(w int) {
			defer wg.Done()
			// open db
			conn := pool.Get(nil)
			defer pool.Put(conn)
			// query
			var nrows int
			err := sqlitex.Exec(conn, utl.SqlSelectBooks, func(stmt *sqlite.Stmt) error {
				nrows++
				var id int
				var name string
				if stmt.ColumnType(0) != sqlite.SQLITE_NULL {
					id = stmt.ColumnInt(0)
				}
				if stmt.ColumnType(1) != sqlite.SQLITE_NULL {
					name = stmt.ColumnText(1)
				}
				if id < 1 || len(name) < 5 {
					log.Fatalf("worker %v: wrong row values", w)
				}
				return nil
			})
			check(err)
			if nrows != nbooks {
				log.Fatalf("worker %v: want %v rows but was %v", w, nbooks, nrows)
			}
		}(w)
	}
	wg.Wait()
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func main() {
	log.Printf("crawshaw")
	utl.ParseFlags()
	benchSimple(utl.Dbfile, utl.Nusers)
	benchComplex(utl.Dbfile, utl.Nprofiles, utl.Ndevices, utl.Nlocations)
	for _, n := range utl.NcarCounts {
		benchMany(utl.Dbfile, n, utl.NcarQueries)
	}
	for _, n := range utl.PlantNameLengths {
		benchLarge(utl.Dbfile, utl.Nplants, utl.NplantQueries, n)
	}
	for _, n := range utl.Ngoroutines {
		benchConcurrent(utl.Dbfile, utl.Nbooks, n)
	}
}
