/*
A benchmark for sqinn-go.
*/
package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cvilsmeier/sqinn-go-bench/utl"
	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func launchOpenPrepare(sqinnPath, dbfile string, clearDbfile bool, sqls []string) *sqinn.Sqinn {
	// make sure db doesn't exist
	if clearDbfile {
		utl.Remove(dbfile)
	}
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbfile)
	check(err)
	// run sqls, if any
	for _, sql := range sqls {
		_, err = sq.ExecOne(sql)
		check(err)
	}
	return sq
}

func closeAndTerminate(sq *sqinn.Sqinn) {
	err := sq.Close()
	check(err)
	err = sq.Terminate()
	check(err)
}

func benchSimple(sqinnPath, dbfile string, nusers int) {
	log.Printf("benchSimple dbfile=%s, nusers=%d", dbfile, nusers)
	// start sqinn
	sq := launchOpenPrepare(sqinnPath, dbfile, true, []string{utl.SqlCreateUsers})
	defer closeAndTerminate(sq)
	// insert users
	tstart := time.Now()
	_, err := sq.ExecOne("BEGIN")
	check(err)
	values := make([]interface{}, 0, nusers*4)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User %d", id)
		age := 33 + i
		rating := 0.13 * float64(i+1)
		values = append(values, id, name, age, rating)
	}
	_, err = sq.Exec(utl.SqlInsertUsers, nusers, 4, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	log.Printf("  insert took %s", utl.Since(tstart))
	tstart = time.Now()
	// query users
	colTypes := []byte{sqinn.ValInt, sqinn.ValText, sqinn.ValInt, sqinn.ValDouble}
	rows, err := sq.Query(utl.SqlSelectUsers, nil, colTypes)
	check(err)
	if len(rows) != nusers {
		log.Fatalf("want %v rows but was %v", nusers, len(rows))
	}
	for _, row := range rows {
		id := row.Values[0].Int.Value
		name := row.Values[1].String.Value
		age := row.Values[2].Int.Value
		rating := row.Values[3].Double.Value
		if id < 1 || len(name) < 5 || age < 33 || rating < 0.13 {
			log.Fatal("wrong row values")
		}
	}
	log.Printf("  query took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchComplex(sqinnPath, dbfile string, nprofiles, ndevices, nlocations int) {
	log.Printf("benchComplex dbfile=%s, nprofiles, ndevices, nlocations = %d, %d, %d", dbfile, nprofiles, ndevices, nlocations)
	// start sqinn
	sq := launchOpenPrepare(sqinnPath, dbfile, true, utl.SqlCreateComplex)
	defer closeAndTerminate(sq)
	// insert profiles
	tstart := time.Now()
	_, err := sq.ExecOne("BEGIN")
	check(err)
	values := make([]interface{}, 0, nprofiles*3)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("Profile %d", p)
		active := p % 2
		values = append(values, profileID, name, active)
	}
	_, err = sq.Exec(utl.SqlInsertProfiles, nprofiles, 3, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// insert devices
	_, err = sq.ExecOne("BEGIN")
	check(err)
	values = make([]interface{}, 0, nprofiles*ndevices*4)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			name := fmt.Sprintf("Device %d %d", p, d)
			active := d % 2
			values = append(values, deviceID, profileID, name, active)
		}
	}
	_, err = sq.Exec(utl.SqlInsertDevices, nprofiles*ndevices, 4, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// insert locations
	_, err = sq.ExecOne("BEGIN")
	check(err)
	values = make([]interface{}, 0, nprofiles*ndevices*nlocations*4)
	for p := 0; p < nprofiles; p++ {
		for d := 0; d < ndevices; d++ {
			deviceID := fmt.Sprintf("device_%d_%d", p, d)
			for l := 0; l < nlocations; l++ {
				locationID := fmt.Sprintf("location_%d_%d_%d", p, d, l)
				name := fmt.Sprintf("Location %d %d %d", p, d, l)
				active := l % 2
				values = append(values, locationID, deviceID, name, active)
			}
		}
	}
	_, err = sq.Exec(utl.SqlInsertLocations, nprofiles*ndevices*nlocations, 4, values)
	check(err)
	_, err = sq.Exec("COMMIT", 1, 0, nil)
	check(err)
	log.Printf("  insert took %s", utl.Since(tstart))
	// query
	tstart = time.Now()
	rows, err := sq.Query(utl.SqlSelectComplex, []interface{}{0, 1}, []byte{sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValInt})
	check(err)
	expectedRows := nprofiles * ndevices * nlocations
	if len(rows) != expectedRows {
		log.Fatalf("  expected %v rows but was %v", expectedRows, len(rows))
	}
	for _, row := range rows {
		locations_id := row.Values[0].String.Value
		locations_deviceId := row.Values[1].String.Value
		locations_name := row.Values[2].String.Value
		locations_active := row.Values[3].Int.Value
		if len(locations_id) < 5 || len(locations_deviceId) < 5 || len(locations_name) < 5 || locations_active < 0 {
			log.Fatalf("wrong row values")
		}
		devices_id := row.Values[4].String.Value
		devices_profileId := row.Values[5].String.Value
		devices_name := row.Values[6].String.Value
		devices_active := row.Values[7].Int.Value
		if len(devices_id) < 5 || len(devices_profileId) < 5 || len(devices_name) < 5 || devices_active < 0 {
			log.Fatalf("wrong row values")
		}
		profiles_id := row.Values[8].String.Value
		profiles_name := row.Values[9].String.Value
		profiles_active := row.Values[10].Int.Value
		if len(profiles_id) < 5 || len(profiles_name) < 5 || profiles_active < 0 {
			log.Fatalf("wrong row values")
		}
	}
	log.Printf("  query took %s", utl.Since(tstart))
	// close and terminate
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchMany(sqinnPath, dbfile string, ncars, nqueries int) {
	log.Printf("benchMany dbfile=%s, ncars=%d, nqueries=%d", dbfile, ncars, nqueries)
	// start sqinn
	sq := launchOpenPrepare(sqinnPath, dbfile, true, []string{utl.SqlCreateCars})
	defer closeAndTerminate(sq)
	// insert cars
	_, err := sq.ExecOne("BEGIN")
	check(err)
	fieldsPerCar := 3 // id, company, model
	values := make([]interface{}, 0, ncars*fieldsPerCar)
	for i := 0; i < ncars; i++ {
		id := i + 1
		company := fmt.Sprintf("Company %d", id)
		model := fmt.Sprintf("Model %d", id)
		values = append(values, id, company, model)
	}
	_, err = sq.Exec(utl.SqlInsertCars, ncars, fieldsPerCar, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		colTypes := []byte{sqinn.ValInt, sqinn.ValText, sqinn.ValText}
		rows, err := sq.Query(utl.SqlSelectCars, nil, colTypes)
		check(err)
		if len(rows) != ncars {
			log.Fatalf("want %v rows but was %v", ncars, len(rows))
		}
		for _, row := range rows {
			id := row.Values[0].Int.Value
			company := row.Values[1].String.Value
			model := row.Values[2].String.Value
			if id < 1 || len(company) < 5 || len(model) < 5 {
				log.Fatal("wrong row values")
			}
		}
	}
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchLarge(sqinnPath, dbfile string, nplants, nqueries, nameLength int) {
	log.Printf("benchLarge dbfile=%s, nplants=%d, nqueries=%d, nameLength=%d", dbfile, nplants, nqueries, nameLength)
	// start sqinn
	sq := launchOpenPrepare(sqinnPath, dbfile, true, []string{utl.SqlCreatePlants})
	defer closeAndTerminate(sq)
	// insert
	name := ""
	for len(name) < nameLength {
		name = name + "Name "
	}
	_, err := sq.ExecOne("BEGIN")
	check(err)
	fieldsPerRow := 2 // id, name
	values := make([]interface{}, 0, nplants*fieldsPerRow)
	for i := 0; i < nplants; i++ {
		id := i + 1
		values = append(values, id, name)
	}
	_, err = sq.Exec(utl.SqlInsertPlants, nplants, fieldsPerRow, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// queries
	tstart := time.Now()
	for i := 0; i < nqueries; i++ {
		colTypes := []byte{sqinn.ValInt, sqinn.ValText}
		rows, err := sq.Query(utl.SqlSelectPlants, nil, colTypes)
		check(err)
		if len(rows) != nplants {
			log.Fatalf("want %v rows but was %v", nplants, len(rows))
		}
		for _, row := range rows {
			id := row.Values[0].Int.Value
			name := row.Values[1].String.Value
			if id < 1 || len(name) < nameLength {
				log.Fatal("wrong row values")
			}
		}
	}
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchConcurrent(sqinnPath, dbfile string, nbooks, nworkers int) {
	log.Printf("benchConcurrent dbfile=%s, nbooks=%d, nworkers=%d", dbfile, nbooks, nworkers)
	// start sqinn
	sq := launchOpenPrepare(sqinnPath, dbfile, true, []string{utl.SqlCreateBooks})
	defer closeAndTerminate(sq)
	// insert
	_, err := sq.ExecOne("BEGIN")
	check(err)
	values := make([]interface{}, 0, nbooks*2)
	for b := 0; b < nbooks; b++ {
		id := b + 1
		name := fmt.Sprintf("Book %d", b)
		values = append(values, id, name)
	}
	_, err = sq.Exec(utl.SqlInsertBooks, nbooks, 2, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// query
	tstart := time.Now()
	var wg sync.WaitGroup
	wg.Add(nworkers)
	for w := 0; w < nworkers; w++ {
		go func(w int) {
			defer wg.Done()
			// start sqinn, do NOT clear db file when starting
			sq := launchOpenPrepare(sqinnPath, dbfile, false, nil)
			defer closeAndTerminate(sq)
			// we have to set busy timeout because many goroutines are accessing
			// the database concurrently and might be stepping on each others toes
			_, err = sq.ExecOne("PRAGMA busy_timeout=10000") // 10s should be enough for everybody
			check(err)
			// query
			rows, err := sq.Query(utl.SqlSelectBooks, nil, []byte{sqinn.ValInt, sqinn.ValText})
			if err != nil {
				log.Fatalf("worker %v err: %v", w, err)
			}
			nrows := len(rows)
			if nrows != nbooks {
				log.Fatalf("worker %v: want %v rows but was %v", w, nbooks, nrows)
			}
			for _, row := range rows {
				id := row.Values[0].Int.Value
				name := row.Values[1].String.Value
				if id < 1 || len(name) < 5 {
					log.Fatalf("worker %v: wrong row values", w)
				}
			}
		}(w)
	}
	wg.Wait()
	log.Printf("  queries took %s", utl.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func main() {
	log.Printf("sqinn")
	sqinnPath := os.Getenv("SQINN_PATH")
	utl.ParseFlags()
	benchSimple(sqinnPath, utl.Dbfile, utl.Nusers)
	benchComplex(sqinnPath, utl.Dbfile, utl.Nprofiles, utl.Ndevices, utl.Nlocations)
	for _, n := range utl.NcarCounts {
		benchMany(sqinnPath, utl.Dbfile, n, utl.NcarQueries)
	}
	for _, n := range utl.PlantNameLengths {
		benchLarge(sqinnPath, utl.Dbfile, utl.Nplants, utl.NplantQueries, n)
	}
	for _, n := range utl.Ngoroutines {
		benchConcurrent(sqinnPath, utl.Dbfile, utl.Nbooks, n)
	}
}
