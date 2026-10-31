package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"smswhiteinfo/customer-api/internal/config"
	"smswhiteinfo/customer-api/internal/httpapi"
	"smswhiteinfo/customer-api/internal/selfcheck"
)

func main() {
	selfcheckFlag := flag.Bool("selfcheck", false, "run pre-flight deployment and environment self-checks")
	flag.Parse()

	if *selfcheckFlag {
		os.Exit(selfcheck.Run())
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers & schedulers (guarded by WORKER_ENABLED and SCHEDULER_ENABLED env flags)
	server.StartBackground(ctx)

	// Graceful shutdown listener
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Termination signal received. Shutting down gracefully...")
		cancel()
		server.StopBackground()
		_ = server.App().ShutdownWithTimeout(30 * time.Second)
	}()

	log.Printf("Customer API listening on %s", applicationConfig.Address)
	if err := server.App().Listen(applicationConfig.Address); err != nil {
		log.Printf("Customer API server stopped: %v", err)
	}
}
