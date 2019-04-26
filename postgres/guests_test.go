package postgres_test

import (
	"checkin/postgres"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //Loads postgres as our database of choice
)

const DEFINE_PATH = "../config/define.sql"

//extracts database definition info from definition sql file
//said file must be formatted with --test comments before an after a continguous block
//of sql statements meant to be executed by the test
func databaseDefinition(t *testing.T) string {
	buf, err := ioutil.ReadFile(DEFINE_PATH)
	if err != nil {
		t.Fatal("Error reading database definition: " + err.Error())
	}

	defineStr := strings.Split(string(buf), "--test")[1]
	return defineStr
}

func initDB(t *testing.T) *sqlx.DB {
	db, err := postgres.Open("postgres://regapp_test:henry@localhost/registrationapp_test")
	if err != nil {
		t.Fatal("Error opening database: " + err.Error())
	}

	_, err = db.Exec(databaseDefinition(t))

	if err != nil {
		t.Fatal("Error adding tables to database: " + err.Error())
	}

	return db
}

func tearDownQuery(db *sqlx.DB, t *testing.T) string {
	var tearDownString string
	rows, err := db.Query(`select 'drop table if exists "' || tablename || '" cascade;' from pg_tables where schemaname = 'public';`)
	if err != nil {
		t.Fatal("Error creating teardown string: " + err.Error())
	}
	for rows.Next() {
		var str string
		rows.Scan(&str)
		tearDownString += str
	}

	return tearDownString
}

func tearDownDB(db *sqlx.DB, t *testing.T) {
	_, err := db.Exec(tearDownQuery(db, t))
	if err != nil {
		t.Fatal("Error dropping all tables: " + err.Error())
	}
}

func TestGuests(t *testing.T) {
	db := initDB(t)

	tearDownDB(db, t)
}
