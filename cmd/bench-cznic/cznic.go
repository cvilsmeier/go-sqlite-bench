package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cvilsmeier/sqinn-go-bench/utl"
	_ "modernc.org/sqlite"
)

func main() {
	utl.Init("cznic")
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func openPrepare(dbfile string, clearDbfile bool, sqls []string) *sql.DB {
	// make sure db doesn't exist
	if clearDbfile {
		utl.Remove(dbfile)
	}
	// open db
	db, err := sql.Open("sqlite", dbfile)
	check(err)
	// prepare schema
	for _, s := range sqls {
		_, err = db.Exec(s)
		check(err)
	}
	return db
}

func closeDb(db *sql.DB) {
	err := db.Close()
	check(err)
}

func benchSimple(dbfile string, nusers int) {
	log.Printf("benchSimple dbfile=%s, nusers=%d", dbfile, nusers)
	db := openPrepare(dbfile, true, []string{utl.SqlCreateUsers})
	defer closeDb(db)
	// insert users
	tstart := time.Now()
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare(utl.SqlInsertUsers)
	check(err)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + id
		rating := 0.13 * float64(id)
		_, err = stmt.Exec(id, name, age, rating)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	log.Printf("  insert took %s", utl.Since(tstart))
	// query users
	tstart = time.Now()
	rows, err := db.Query(utl.SqlSelectUsers)
	check(err)
	nrows := 0
	var id sql.NullInt32
	var name sql.NullString
	var age sql.NullInt32
	var rating sql.NullFloat64
	for rows.Next() {
		nrows++
		err = rows.Scan(&id, &name, &age, &rating)
		check(err)
		if id.Int32 < 1 || len(name.String) < 5 || age.Int32 < 33 || rating.Float64 < 0.13 {
			log.Fatal("wrong row values")
		}
	}
	if nrows != nusers {
		log.Fatalf("expected %v rows but was %v", nusers, nrows)
	}
	log.Printf("  query took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchComplex(dbfile string, nprofiles, ndevices, nlocations int) {
	log.Printf("benchComplex dbfile=%s, nprofiles, ndevices, nlocations = %d, %d, %d", dbfile, nprofiles, ndevices, nlocations)
	db := openPrepare(dbfile, true, utl.SqlCreateComplex)
	defer closeDb(db)
	// insert profiles
	tstart := time.Now()
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare(utl.SqlInsertProfiles)
	check(err)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("Profile %d", p)
		active := p % 2
		_, err = stmt.Exec(profileID, name, active)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	// insert devices
	tx, err = db.Begin()
	check(err)
	stmt, err = tx.Prepare(utl.SqlInsertDevices)
	check(err)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			name := fmt.Sprintf("Device %d %d", p, d)
			active := d % 2
			_, err = stmt.Exec(deviceID, profileID, name, active)
			check(err)
		}
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	// insert locations
	tx, err = db.Begin()
	check(err)
	stmt, err = tx.Prepare(utl.SqlInsertLocations)
	check(err)
	for p := 0; p < nprofiles; p++ {
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			for l := 0; l < nlocations; l++ {
				locationID := fmt.Sprintf("location_%d_%d_%d", p, d, l)
				name := fmt.Sprintf("Location %d %d %d", p, d, l)
				active := l % 2
				_, err = stmt.Exec(locationID, deviceID, name, active)
				check(err)
			}
		}
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	log.Printf("  insert took %s", utl.Since(tstart))
	// query
	tstart = time.Now()
	rows, err := db.Query(utl.SqlSelectComplex, 0, 1)
	check(err)
	nrows := 0
	var locations_id sql.NullString
	var locations_deviceId sql.NullString
	var locations_name sql.NullString
	var locations_active sql.NullInt32
	var devices_id sql.NullString
	var devices_profileId sql.NullString
	var devices_name sql.NullString
	var devices_active sql.NullInt32
	var profiles_id sql.NullString
	var profiles_name sql.NullString
	var profiles_active sql.NullInt32
	for rows.Next() {
		nrows++
		rows.Scan(
			&locations_id,
			&locations_deviceId,
			&locations_name,
			&locations_active,
			&devices_id,
			&devices_profileId,
			&devices_name,
			&devices_active,
			&profiles_id,
			&profiles_name,
			&profiles_active,
		)
		if len(locations_id.String) < 5 || len(locations_deviceId.String) < 5 || len(locations_name.String) < 5 || locations_active.Int32 < 0 {
			log.Fatalf("wrong row values")
		}
		if len(devices_id.String) < 5 || len(devices_profileId.String) < 5 || len(devices_name.String) < 5 || devices_active.Int32 < 0 {
			log.Fatalf("wrong row values")
		}
		if len(profiles_id.String) < 5 || len(profiles_name.String) < 5 || profiles_active.Int32 < 0 {
			log.Fatalf("wrong row values")
		}
	}
	expectedRows := nprofiles * ndevices * nlocations
	if nrows != expectedRows {
		log.Fatalf("expected %v rows but was %v", expectedRows, nrows)
	}
	// done
	log.Printf("  query took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchMany(dbfile string, ncars, nqueries int) {
	log.Printf("benchMany dbfile=%s, ncars=%d, nqueries=%d", dbfile, ncars, nqueries)
	db := openPrepare(dbfile, true, []string{utl.SqlCreateCars})
	defer closeDb(db)
	// insert
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare(utl.SqlInsertCars)
	check(err)
	for i := 0; i < ncars; i++ {
		id := i + 1
		company := fmt.Sprintf("Company %d", id)
		model := fmt.Sprintf("Model %d", id)
		_, err = stmt.Exec(id, company, model)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		rows, err := db.Query(utl.SqlSelectCars)
		check(err)
		nrows := 0
		var id sql.NullInt32
		var company sql.NullString
		var model sql.NullString
		for rows.Next() {
			nrows++
			rows.Scan(&id, &company, &model)
			if id.Int32 < 1 || len(company.String) < 5 || len(model.String) < 5 {
				log.Fatal("wrong row values")
			}
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
	db := openPrepare(dbfile, true, []string{utl.SqlCreatePlants})
	defer closeDb(db)
	// insert
	name := ""
	for len(name) < nameLength {
		name = name + "Name "
	}
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare(utl.SqlInsertPlants)
	check(err)
	for i := 0; i < nplants; i++ {
		id := i + 1
		_, err = stmt.Exec(id, name)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		rows, err := db.Query(utl.SqlSelectPlants)
		check(err)
		nrows := 0
		var id sql.NullInt32
		var name sql.NullString
		for rows.Next() {
			nrows++
			rows.Scan(&id, &name)
			if id.Int32 < 1 || len(name.String) < nameLength {
				log.Fatal("wrong row values")
			}
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
	db := openPrepare(dbfile, true, []string{utl.SqlCreateBooks})
	defer closeDb(db)
	// insert
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare(utl.SqlInsertBooks)
	check(err)
	for b := 0; b < nbooks; b++ {
		id := b + 1
		name := fmt.Sprintf("Book %d", id)
		_, err = stmt.Exec(id, name)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	// query
	tstart := time.Now()
	var wg sync.WaitGroup
	wg.Add(nworkers)
	for w := 0; w < nworkers; w++ {
		go func(w int) {
			defer wg.Done()
			db := openPrepare(dbfile, false, nil)
			defer closeDb(db)
			rows, err := db.Query(utl.SqlSelectBooks)
			if err != nil {
				log.Fatalf("worker %v: %v", w, err)
			}
			nrows := 0
			var id sql.NullInt32
			var name sql.NullString
			for rows.Next() {
				nrows++
				rows.Scan(&id, &name)
				if id.Int32 < 1 || len(name.String) < 5 {
					log.Fatalf("worker %v: wrong row values", w)
				}
			}
			if nrows != nbooks {
				log.Fatalf("worker %v: want %v rows but was %v", w, nbooks, nrows)
			}
		}(w)
	}
	wg.Wait()
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}
