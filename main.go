package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/livingstone23/gosimplebank/api"
	db "github.com/livingstone23/gosimplebank/db/sqlc"
	"github.com/livingstone23/gosimplebank/util"
	"log"
)

/*
const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)
*/

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	fmt.Println("Start SimpleBank")

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err.Error())
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}

}
