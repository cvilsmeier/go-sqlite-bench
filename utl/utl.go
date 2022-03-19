package utl

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	// benchSimple
	SqlCreateUsers = "CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)"
	SqlInsertUsers = "INSERT INTO users (id, name, age, rating) VALUES (?, ?, ?, ?)"
	SqlSelectUsers = "SELECT id, name, age, rating FROM users ORDER BY id"
	// benchComplex
	SqlCreateComplex = []string{
		"PRAGMA foreign_keys=1",
		"CREATE TABLE profiles (id VARCHAR PRIMARY KEY NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL)",
		"CREATE INDEX idx_profiles_name ON profiles(name);",
		"CREATE INDEX idx_profiles_active ON profiles(active);",
		"CREATE TABLE devices (id VARCHAR PRIMARY KEY NOT NULL, profileId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (profileId) REFERENCES profiles(id))",
		"CREATE INDEX idx_devices_profileId ON devices(profileId);",
		"CREATE INDEX idx_devices_name ON devices(name);",
		"CREATE INDEX idx_devices_active ON devices(active);",
		"CREATE TABLE locations (id VARCHAR PRIMARY KEY NOT NULL, deviceId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (deviceId) REFERENCES devices(id))",
		"CREATE INDEX idx_locations_userId ON locations(deviceId);",
		"CREATE INDEX idx_locations_name ON locations(name);",
		"CREATE INDEX idx_locations_active ON locations(active);",
	}
	SqlInsertProfiles  = "INSERT INTO profiles (id, name, active) VALUES (?, ?, ?)"
	SqlInsertDevices   = "INSERT INTO devices (id, profileId, name, active) VALUES (?, ?, ?, ?)"
	SqlInsertLocations = "INSERT INTO locations (id, deviceId, name, active) VALUES (?, ?, ?, ?)"
	SqlSelectComplex   = "SELECT " +
		" locations.id, locations.deviceId, locations.name, locations.active, " +
		" devices.id, devices.profileId, devices.name, devices.active, " +
		" profiles.id, profiles.name, profiles.active " +
		"FROM locations " +
		"LEFT JOIN devices ON devices.id = locations.deviceId " +
		"LEFT JOIN profiles ON profiles.id = devices.profileId " +
		"WHERE locations.active = ? OR locations.active = ? " +
		"ORDER BY locations.name, locations.id, devices.name, devices.id, profiles.name, profiles.id"
	// benchMany
	SqlCreateCars = "CREATE TABLE cars (id INTEGER PRIMARY KEY NOT NULL, company VARCHAR, model VARCHAR)"
	SqlInsertCars = "INSERT INTO cars (id, company, model) VALUES (?, ?, ?)"
	SqlSelectCars = "SELECT id, company, model FROM cars ORDER BY id"
	// benchLarge
	SqlCreatePlants = "CREATE TABLE plants (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)"
	SqlInsertPlants = "INSERT INTO plants (id, name) VALUES (?, ?)"
	SqlSelectPlants = "SELECT id, name FROM plants ORDER BY id"
	// benchConcurrent
	SqlCreateBooks = "CREATE TABLE books (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)"
	SqlInsertBooks = "INSERT INTO books (id,name) VALUES (?,?)"
	SqlSelectBooks = "SELECT id, name FROM books ORDER BY id"
)

var (
	Nusers           = 1000 * 1000             // benchSimple
	Nprofiles        = 200                     // benchComplex
	Ndevices         = 100                     // benchComplex
	Nlocations       = 10                      // benchComplex
	NcarCounts       = []int{10, 100, 1000}    // benchMany
	NcarQueries      = 1000                    // benchMany
	Nplants          = 500                     // benchLarge
	NplantQueries    = 100                     // benchLarge
	PlantNameLengths = []int{2000, 4000, 8000} // benchLarge
	Nbooks           = 1000 * 1000             // benchConcurrent
	Ngoroutines      = []int{2, 4, 8}          // benchConcurrent
)

var (
	Dbfile = "./sqinn-go-bench.db"
)

func Init(name string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix(name + ": ")
	flag.StringVar(&Dbfile, "db", Dbfile, "path to db file for test")
	flag.Parse()
}

func Remove(dbfile string) {
	if dbfile == "" {
		return
	}
	if strings.Contains(dbfile, ":memory:") {
		return
	}
	// best effort, ignore errors
	err := os.Remove(dbfile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	err = os.Remove(dbfile + "-journal")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	err = os.Remove(dbfile + "-wal")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	err = os.Remove(dbfile + "-shm")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

func Since(t time.Time) string {
	return fmt.Sprintf("%d ms", time.Since(t).Milliseconds())
}

func Fsize(name string) string {
	fi, err := os.Stat(name)
	if err != nil {
		return fmt.Sprintf("fsize: %s", err)
	}
	return fmt.Sprintf("%d (%d K)", fi.Size(), fi.Size()/int64(1024))
}
