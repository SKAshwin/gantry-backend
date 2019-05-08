package postgres_test

//This file is for seting up and tearing down the test database
//TestMain runs before/after every test in the suite

import (
	"checkin/postgres"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //Loads postgres as our database of choice
)

const DEFINE_PATH = "../config/define.sql"
const TEST_DATA_PATH = "../config/testData.sql"

var db *sqlx.DB

//extracts database definition info from definition sql file
//said file must be formatted with --test comments before an after a continguous block
//of sql statements meant to be executed by the test
func databaseDefinition() string {
	buf, err := ioutil.ReadFile(DEFINE_PATH)
	if err != nil {
		log.Fatal("Error reading database definition: " + err.Error())
	}

	defineStr := strings.Split(string(buf), "--test")[1]
	return defineStr
}

func initDB() *sqlx.DB {
	db, err := postgres.Open("postgres://regapp_test:henry@localhost/registrationapp_test")
	tearDownDB(db) //tear down any remnants of the DB in case last test was not cleaned up properly
	if err != nil {
		log.Fatal("Error opening database: " + err.Error())
	}

	_, err = db.Exec(databaseDefinition())

	if err != nil {
		log.Fatal("Error adding tables to database: " + err.Error())
	}

	return db
}

func tearDownQuery(db *sqlx.DB) string {
	var tearDownString string
	rows, err := db.Query(`select 'drop table if exists "' || tablename || '" cascade;' from pg_tables where schemaname = 'public';`)
	if err != nil {
		log.Fatal("Error creating teardown string: " + err.Error())
	}
	for rows.Next() {
		var str string
		rows.Scan(&str)
		tearDownString += str
	}

	return tearDownString
}

func loadTestData(db *sqlx.DB) error {
	_, err := db.Exec(testData())
	return err
}

func testData() string {
	buf, err := ioutil.ReadFile(TEST_DATA_PATH)
	if err != nil {
		log.Fatal("Error reading database definition: " + err.Error())
	}
	return string(buf)
}

func tearDownDB(db *sqlx.DB) {
	_, err := db.Exec(tearDownQuery(db))
	if err != nil {
		log.Fatal("Error dropping all tables: " + err.Error())
	}
}

func TestMain(m *testing.M) {
	db = initDB()
	defer func() {
		//if a panic occurs, make sure to tear down the DB
		if r := recover(); r != nil {
			tearDownDB(db)
			log.Println("Recovered from panic to tear down database, resuming panic")
			panic(r)
		}
	}()
	err := loadTestData(db)
	if err != nil {
		log.Println("Error loading test data: " + err.Error())
		tearDownDB(db)
		log.Fatal()
	} else {
		retCode := m.Run()
		tearDownDB(db)
		os.Exit(retCode)
	}
}
