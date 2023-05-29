package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/livingstone23/gosimplebank/util"
	"log"
	"os"
	"testing"
)

/*
const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)
*/

var testQueries *Queries
var testDB *sql.DB

// TestMain, take testing.M object as input
// is the main entry point
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config: ", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err.Error())
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
