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

func benchSimple(sqinnPath, dbfile string, nusers int) {
	log.Printf("benchSimple dbfile=%s, nusers=%d", dbfile, nusers)
	// make sure db doesn't exist
	utl.Remove(dbfile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbfile)
	check(err)
	// prepare schema
	_, err = sq.ExecOne(utl.SqlCreateUsers)
	check(err)
	// insert users
	tstart := time.Now()
	_, err = sq.ExecOne("BEGIN")
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
	log.Printf("  insert took %s", time.Since(tstart))
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
	log.Printf("  query took %s", time.Since(tstart))
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchComplex(sqinnPath, dbfile string, nprofiles, ndevices, nlocations int) {
	log.Printf("benchComplex dbfile=%s, nprofiles, ndevices, nlocations = %d, %d, %d", dbfile, nprofiles, ndevices, nlocations)
	// make sure db doesn't exist
	utl.Remove(dbfile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbfile)
	check(err)
	// prepare schema
	for _, sql := range utl.SqlCreateComplex {
		_, err = sq.ExecOne(sql)
		check(err)
	}
	// insert profiles
	tstart := time.Now()
	_, err = sq.ExecOne("BEGIN")
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
	log.Printf("  insert took %s", time.Since(tstart))
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
	log.Printf("  query took %s", time.Since(tstart))
	// close and terminate
	err = sq.Close()
	check(err)
	err = sq.Terminate()
	check(err)
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func benchConcurrent(sqinnPath, dbfile string, nbooks, nworkers int) {
	log.Printf("benchConcurrent dbfile=%s, nbooks=%d, nworkers=%d", dbfile, nbooks, nworkers)
	// make sure db doesn't exist
	utl.Remove(dbfile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	defer sq.Terminate()
	// open db
	err = sq.Open(dbfile)
	check(err)
	defer sq.Close()
	// prepare schema
	_, err = sq.ExecOne(utl.SqlCreateBooks)
	check(err)
	// insert
	_, err = sq.ExecOne("BEGIN")
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
			// launch sqinn
			sq, err := sqinn.Launch(sqinn.Options{
				SqinnPath: sqinnPath,
			})
			check(err)
			defer sq.Terminate()
			// open db
			err = sq.Open(dbfile)
			check(err)
			defer sq.Close()
			// we have to set busy timeout, sqinn does not support the unlock-notify API
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
	// done
	log.Printf("  queries took %s", time.Since(tstart))
	log.Printf("  fsize %v", utl.Fsize(dbfile))
}

func main() {
	sqinnPath := os.Getenv("SQINN_PATH")
	log.Printf("sqinn %s", sqinnPath)
	utl.ParseFlags()
	benchSimple(sqinnPath, utl.Dbfile, utl.Nusers)
	benchComplex(sqinnPath, utl.Dbfile, utl.Nprofiles, utl.Ndevices, utl.Nlocations)
	benchConcurrent(sqinnPath, utl.Dbfile, utl.Nbooks, 2)
	benchConcurrent(sqinnPath, utl.Dbfile, utl.Nbooks, 4)
	benchConcurrent(sqinnPath, utl.Dbfile, utl.Nbooks, 8)
}
