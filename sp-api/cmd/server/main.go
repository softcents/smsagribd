package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"smswhiteinfo/sp-api/internal/config"
	"smswhiteinfo/sp-api/internal/httpapi"
)

func main() {
	applicationConfig, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := sql.Open("mysql", applicationConfig.Database.FormatDSN())
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
	log.Printf("SP API listening on %s", applicationConfig.Address)
	log.Fatal(server.App().Listen(applicationConfig.Address))
}
