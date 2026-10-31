// Package selfcheck implements deployment self-check logic for Customer API.
package selfcheck

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"smswhiteinfo/customer-api/internal/crypto"
)

type result struct {
	name    string
	ok      bool
	message string
}

func pass(name, msg string) result { return result{name: name, ok: true, message: msg} }
func fail(name, msg string) result { return result{name: name, ok: false, message: msg} }

func printResult(r result) {
	label := "[PASS]"
	if !r.ok {
		label = "[FAIL]"
	}
	fmt.Printf("  %s %s: %s\n", label, r.name, r.message)
}

var requiredTables = []string{
	"users", "messages", "batches", "senders", "gateways",
	"deposits", "transactions", "payment_webhooks", "access_tokens",
	"ip_whitelist_requests", "verification_otps", "cache",
}

var spotCheckColumns = []struct{ table, column string }{
	{"messages", "gateway_message_id"},
	{"messages", "poll_attempt"},
	{"messages", "next_poll_at"},
	{"users", "balance"},
	{"deposits", "processed_at"},
}

var requiredStorageDirs = []string{"ids", "live_photos", "trade_licenses", "batches"}

func checkDBConnection(db *sql.DB) result {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	dbName := os.Getenv("DB_DATABASE")

	if err := db.PingContext(context.Background()); err != nil {
		return fail("DB_CONNECTION", fmt.Sprintf("Ping failed: %v", err))
	}
	return pass("DB_CONNECTION", fmt.Sprintf("Connected to %s@%s:%s", dbName, host, port))
}

func checkDBTables(db *sql.DB) result {
	ctx := context.Background()
	var missing []string
	for _, tbl := range requiredTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", tbl).Scan(&count)
		if err != nil || count == 0 {
			missing = append(missing, tbl)
		}
	}
	if len(missing) > 0 {
		return fail("DB_TABLES", fmt.Sprintf("Missing tables: %s", strings.Join(missing, ", ")))
	}
	return pass("DB_TABLES", fmt.Sprintf("All %d required tables present", len(requiredTables)))
}

func checkDBColumns(db *sql.DB) result {
	ctx := context.Background()
	var notes []string
	var missing []string

	for _, col := range spotCheckColumns {
		var count int
		err := db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
			col.table, col.column,
		).Scan(&count)
		if err != nil || count == 0 {
			ref := fmt.Sprintf("%s.%s", col.table, col.column)
			if ref == "deposits.processed_at" {
				notes = append(notes, "deposits.processed_at not present (deposits uses updated_at/payment_webhooks)")
			} else {
				missing = append(missing, ref)
			}
		}
	}

	if len(missing) > 0 {
		return fail("DB_COLUMNS", fmt.Sprintf("Missing columns: %s", strings.Join(missing, ", ")))
	}
	msg := fmt.Sprintf("Spot-checked %d key columns", len(spotCheckColumns))
	if len(notes) > 0 {
		msg += " (" + strings.Join(notes, "; ") + ")"
	}
	return pass("DB_COLUMNS", msg)
}

func checkAppKeyDecrypt(db *sql.DB) []result {
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		return []result{fail("APP_KEY_DECRYPT", "APP_KEY env var is not set")}
	}

	enc, err := crypto.NewLaravelEncrypter(appKey)
	if err != nil {
		return []result{fail("APP_KEY_DECRYPT", fmt.Sprintf("Invalid APP_KEY format: %v", err))}
	}

	rows, err := db.QueryContext(context.Background(), "SELECT id, driver, credentials FROM gateways ORDER BY id ASC")
	if err != nil {
		return []result{fail("APP_KEY_DECRYPT", fmt.Sprintf("Cannot query gateways: %v", err))}
	}
	defer rows.Close()

	var results []result
	gatewayCount := 0
	for rows.Next() {
		gatewayCount++
		var id int64
		var driver string
		var creds sql.NullString
		if err := rows.Scan(&id, &driver, &creds); err != nil {
			results = append(results, fail("APP_KEY_DECRYPT", fmt.Sprintf("Row scan error: %v", err)))
			continue
		}

		if !creds.Valid || creds.String == "" || creds.String == "{}" {
			results = append(results, pass(
				fmt.Sprintf("APP_KEY_DECRYPT [gw %d %s]", id, driver),
				"empty credentials (no decryption needed)",
			))
			continue
		}

		credsVal := creds.String
		if !strings.HasPrefix(credsVal, "eyJ") {
			results = append(results, pass(
				fmt.Sprintf("APP_KEY_DECRYPT [gw %d %s]", id, driver),
				"plain JSON credentials (unencrypted)",
			))
			continue
		}

		_, decErr := enc.DecryptString(credsVal)
		if decErr != nil {
			results = append(results, fail(
				fmt.Sprintf("APP_KEY_DECRYPT [gw %d %s]", id, driver),
				fmt.Sprintf("DECRYPT_FAILED: %v", decErr),
			))
		} else {
			results = append(results, pass(
				fmt.Sprintf("APP_KEY_DECRYPT [gw %d %s]", id, driver),
				"DECRYPT_OK",
			))
		}
	}

	if gatewayCount == 0 {
		results = append(results, pass("APP_KEY_DECRYPT", "No gateways configured in DB"))
	}
	return results
}

func checkStorageFolders() result {
	base := os.Getenv("STORAGE_PATH")
	if base == "" {
		for _, candidate := range []string{
			"./storage/app/private",
			"../storage/app/private",
			"/var/www/storage/app/private",
		} {
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				base = candidate
				break
			}
		}
	}
	if base == "" {
		base = "./storage/app/private"
	}

	fi, err := os.Stat(base)
	if err != nil || !fi.IsDir() {
		return fail("STORAGE_FOLDERS", fmt.Sprintf("Private storage path not found: %s", base))
	}

	testFile := filepath.Join(base, ".selfcheck_write_test")
	if err := os.WriteFile(testFile, []byte("ok"), 0600); err != nil {
		return fail("STORAGE_FOLDERS", fmt.Sprintf("%s is not writable: %v", base, err))
	}
	_ = os.Remove(testFile)

	var missing []string
	for _, sub := range requiredStorageDirs {
		subPath := filepath.Join(base, sub)
		if si, err := os.Stat(subPath); err != nil || !si.IsDir() {
			missing = append(missing, sub+"/")
		}
	}
	if len(missing) > 0 {
		return fail("STORAGE_FOLDERS", fmt.Sprintf("%s exists but sub-dirs missing: %s", base, strings.Join(missing, " ")))
	}
	return pass("STORAGE_FOLDERS", fmt.Sprintf("%s exists, %s OK",
		base, strings.Join(requiredStorageDirs, "/ ")+"/",
	))
}

func checkSMTPConfig() result {
	var missing []string
	for _, key := range []string{"MAIL_HOST", "MAIL_PORT", "MAIL_FROM_ADDRESS"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fail("SMTP_CONFIG", fmt.Sprintf("%s is not set (optional fallback will log emails)", strings.Join(missing, ", ")))
	}
	return pass("SMTP_CONFIG", fmt.Sprintf("MAIL_HOST=%s MAIL_PORT=%s", os.Getenv("MAIL_HOST"), os.Getenv("MAIL_PORT")))
}

func checkPiprapayConfig() result {
	var missing []string
	for _, key := range []string{"PIPRAPAY_API_KEY", "PIPRAPAY_BASE_URL"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fail("PIPRAPAY_CONFIG", fmt.Sprintf("%s not set", strings.Join(missing, ", ")))
	}
	return pass("PIPRAPAY_CONFIG", "Keys present (PIPRAPAY_API_KEY, PIPRAPAY_BASE_URL)")
}

// Run executes the complete suite of pre-flight environment checks and returns 0 on success, 1 on failure.
func Run() int {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	fmt.Println("=== SMS Whiteinfo BD Pre-Flight Deployment Self-Check ===")
	fmt.Println()

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USERNAME")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("  [FAIL] DB_CONNECTION: Cannot open DB: %v\n", err)
		return 1
	}
	defer db.Close()

	var allResults []result

	allResults = append(allResults, checkDBConnection(db))
	allResults = append(allResults, checkDBTables(db))
	allResults = append(allResults, checkDBColumns(db))
	allResults = append(allResults, checkAppKeyDecrypt(db)...)
	allResults = append(allResults, checkStorageFolders())
	allResults = append(allResults, checkSMTPConfig())
	allResults = append(allResults, checkPiprapayConfig())

	fmt.Println()
	var failCount, passCount int
	for _, r := range allResults {
		printResult(r)
		if r.ok {
			passCount++
		} else {
			failCount++
		}
	}
	fmt.Println()

	if failCount > 0 {
		fmt.Printf("  RESULT: %d check(s) FAILED, %d passed\n\n", failCount, passCount)
		return 1
	}
	fmt.Printf("  RESULT: All %d checks PASSED\n\n", passCount)
	return 0
}
