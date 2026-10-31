package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	minutes := flag.Int("minutes", 30, "Age threshold in minutes for stale records")
	dryRun := flag.Bool("dry-run", true, "Print stale records without making changes (default true)")
	execute := flag.Bool("execute", false, "Actually fix stale records (requires --dry-run=false)")
	flag.Parse()

	// If --execute is passed, treat as non-dry-run
	if *execute {
		*dryRun = false
	}

	mode := "dry-run"
	if !*dryRun {
		mode = "execute"
	}

	// Load .env files
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	// Build DSN from env
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_DATABASE", "sms_panel")
	dbUser := getEnv("DB_USERNAME", "root")
	dbPass := getEnv("DB_PASSWORD", "")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	fmt.Printf("STALE RECORDS (threshold: %d min, mode: %s)\n", *minutes, mode)

	type staleRow struct {
		userID int64
		count  int
	}

	totalStale := 0
	affectedUsers := map[int64]bool{}

	// 1. Pending messages older than N minutes
	pendingMsgs := queryStale(db, `
		SELECT user_id, COUNT(*) as cnt
		FROM messages
		WHERE status = 'pending'
		  AND created_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
		GROUP BY user_id
	`, *minutes)
	for _, r := range pendingMsgs {
		fmt.Printf("[messages/pending]   user_id=%d: %d rows\n", r.userID, r.count)
		totalStale += r.count
		affectedUsers[r.userID] = true
	}

	// 2. Sending messages older than N minutes
	sendingMsgs := queryStale(db, `
		SELECT user_id, COUNT(*) as cnt
		FROM messages
		WHERE status = 'sending'
		  AND updated_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
		GROUP BY user_id
	`, *minutes)
	for _, r := range sendingMsgs {
		fmt.Printf("[messages/sending]   user_id=%d: %d rows\n", r.userID, r.count)
		totalStale += r.count
		affectedUsers[r.userID] = true
	}

	// 3. Processing batches older than N minutes
	processingBatches := queryStale(db, `
		SELECT user_id, COUNT(*) as cnt
		FROM batches
		WHERE status = 'processing'
		  AND updated_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
		GROUP BY user_id
	`, *minutes)
	for _, r := range processingBatches {
		fmt.Printf("[batches/processing] user_id=%d: %d rows\n", r.userID, r.count)
		totalStale += r.count
		affectedUsers[r.userID] = true
	}

	// 4. Pending batches older than N minutes
	pendingBatches := queryStale(db, `
		SELECT user_id, COUNT(*) as cnt
		FROM batches
		WHERE status = 'pending'
		  AND created_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
		GROUP BY user_id
	`, *minutes)
	for _, r := range pendingBatches {
		fmt.Printf("[batches/pending]    user_id=%d: %d rows\n", r.userID, r.count)
		totalStale += r.count
		affectedUsers[r.userID] = true
	}

	fmt.Printf("Total stale: %d records across %d users\n", totalStale, len(affectedUsers))

	if totalStale == 0 {
		os.Exit(0)
	}

	// Execute fixes if requested
	if !*dryRun {
		// Reset "sending" messages > N min back to "pending"
		resetGroups := queryStale(db, `
			SELECT user_id, COUNT(*) as cnt
			FROM messages
			WHERE status = 'sending'
			  AND updated_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
			GROUP BY user_id
		`, *minutes)

		for _, r := range resetGroups {
			res, err := db.Exec(`
				UPDATE messages
				SET status = 'pending', updated_at = NOW()
				WHERE status = 'sending'
				  AND updated_at < DATE_SUB(NOW(), INTERVAL ? MINUTE)
				  AND user_id = ?
			`, *minutes, r.userID)
			if err != nil {
				log.Printf("[EXECUTE] Failed to reset sending messages for user_id=%d: %v", r.userID, err)
				continue
			}
			affected, _ := res.RowsAffected()
			log.Printf("[EXECUTE] Reset %d sending->pending messages for user_id=%d", affected, r.userID)
		}
	}

	os.Exit(1)
}

type staleRow struct {
	userID int64
	count  int
}

func queryStale(db *sql.DB, query string, minutes int) []staleRow {
	rows, err := db.Query(query, minutes)
	if err != nil {
		log.Printf("Query error: %v", err)
		return nil
	}
	defer rows.Close()

	var results []staleRow
	for rows.Next() {
		var r staleRow
		if err := rows.Scan(&r.userID, &r.count); err == nil {
			results = append(results, r)
		}
	}
	return results
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
