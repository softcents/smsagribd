package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"smswhiteinfo/superman-api/internal/config"
	"smswhiteinfo/superman-api/internal/httpapi"
)

func main() {
	applicationConfig, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := sql.Open(applicationConfig.DBDriver, applicationConfig.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}

	server := httpapi.New(applicationConfig, database)
	log.Printf("Superman API listening on %s", applicationConfig.Address)
	log.Fatal(server.App().Listen(applicationConfig.Address))
}
