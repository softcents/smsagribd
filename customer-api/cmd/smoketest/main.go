package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"smswhiteinfo/customer-api/internal/crypto"
	"smswhiteinfo/customer-api/internal/sms"
	"smswhiteinfo/customer-api/internal/sms/drivers"
)

func main() {
	gatewayID := flag.Int64("gateway-id", 0, "Gateway ID to send through")
	to := flag.String("to", "", "Recipient phone number (e.g. 01700000000 or 8801700000000)")
	message := flag.String("message", "Test SMS from SMS-WhiteInfo staging", "Message body")
	from := flag.String("from", "", "Sender ID / CLI")
	confirm := flag.Bool("i-understand-this-sends-a-real-sms", false, "Required confirmation flag to allow outbound SMS")

	flag.Parse()

	if !*confirm {
		fmt.Println("================================================================================")
		fmt.Println("REFUSED: This tool sends a REAL SMS through a live carrier gateway.")
		fmt.Println("To proceed, you must explicitly pass: --i-understand-this-sends-a-real-sms")
		fmt.Println("Example:")
		fmt.Println("  go run ./cmd/smoketest --gateway-id=5 --to=88017XXXXXXXX --i-understand-this-sends-a-real-sms")
		fmt.Println("================================================================================")
		os.Exit(1)
	}

	if *gatewayID == 0 || *to == "" {
		log.Fatalf("Error: --gateway-id and --to are required.")
	}

	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
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
		pass := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_DATABASE")
		if dbname == "" {
			dbname = "sms_panel"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)
	}

	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		log.Fatalf("Error: APP_KEY environment variable is required to decrypt gateway credentials.")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var gwName, gwDriver, gwCreds string
	var isEnabled bool
	err = db.QueryRowContext(ctx, `
		SELECT name, driver, credentials, is_enabled
		FROM gateways
		WHERE id = ?
	`, *gatewayID).Scan(&gwName, &gwDriver, &gwCreds, &isEnabled)
	if err != nil {
		log.Fatalf("Failed to find gateway #%d: %v", *gatewayID, err)
	}

	enc, err := crypto.NewLaravelEncrypter(appKey)
	if err != nil {
		log.Fatalf("Failed to initialize encrypter: %v", err)
	}

	creds, err := enc.DecryptArray(gwCreds)
	if err != nil {
		log.Fatalf("Failed to decrypt gateway credentials: %v", err)
	}

	driver, err := drivers.CreateDriver(gwDriver, creds)
	if err != nil {
		log.Fatalf("Failed to create gateway driver %q: %v", gwDriver, err)
	}

	// Sanitized request preview
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Printf("SMOKE TEST GATEWAY #%d: %s (%s)\n", *gatewayID, gwName, gwDriver)
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Printf("Recipient:   %s\n", *to)
	fmt.Printf("Sender CLI:  %s\n", *from)
	fmt.Printf("Message:     %s\n", *message)
	fmt.Printf("Credentials: [REDACTED — %d keys present: %s]\n", len(creds), strings.Join(keysOf(creds), ", "))
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Println("Dispatching outbound SMS request...")

	req := sms.SmsRequest{
		To:        *to,
		Body:      *message,
		From:      *from,
		IsUnicode: false,
	}

	start := time.Now()
	resp, err := driver.Send(ctx, req)
	latency := time.Since(start)

	if err != nil {
		fmt.Printf("RESULT: FAILED (%v) in %v\n", err, latency)
		os.Exit(2)
	}

	msgID := ""
	if resp.GatewayMessageId != nil {
		msgID = *resp.GatewayMessageId
	}
	fmt.Printf("RESULT: SUCCESS (HTTP latency: %v)\n", latency)
	fmt.Printf("Gateway Message ID: %s\n", msgID)
	fmt.Printf("Raw Response Data:  %v\n", resp.Raw)
	fmt.Println("─────────────────────────────────────────────────────────────────")
}

func keysOf(m map[string]any) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
