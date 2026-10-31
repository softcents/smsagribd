package batch

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/sms_panel_scratch?parseTime=true")
	if err != nil {
		t.Skip("Scratch DB not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("Scratch DB not pingable")
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return db
}

func setupTestBatchEnvironment(t *testing.T) (*sql.DB, int64, int64, string, func()) {
	db := getTestDB(t)
	ctx := context.Background()

	now := time.Now().UnixNano()
	tmpDir, err := os.MkdirTemp("", "batch_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// 1. Create user with enough balance
	uRes, err := db.ExecContext(ctx, `
		INSERT INTO users (name, email, phone, password, balance, type, status, sms_auto_approved, created_at, updated_at)
		VALUES ('Batch Test User', ?, ?, 'hash', 1000.00, 'customer', 'active', 1, NOW(), NOW())
	`, fmt.Sprintf("batch_u_%d@test.com", now), fmt.Sprintf("+88017%08d", now%100000000))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID, _ := uRes.LastInsertId()

	// 2. Create sender
	sRes, err := db.ExecContext(ctx, `
		INSERT INTO senders (sender_id, type, is_enabled, created_at, updated_at)
		VALUES (?, 'non_masking', 1, NOW(), NOW())
	`, fmt.Sprintf("CLI_%d", now%100000))
	if err != nil {
		t.Fatalf("Failed to create sender: %v", err)
	}
	senderID, _ := sRes.LastInsertId()

	teardown := func() {
		_ = os.RemoveAll(tmpDir)
		_, _ = db.ExecContext(ctx, "DELETE FROM messages WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM batches WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM senders WHERE id = ?", senderID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
		db.Close()
	}

	return db, userID, senderID, tmpDir, teardown
}

func TestBatchProcessor_Manual_And_Dedupe_And_Invalid(t *testing.T) {
	db, userID, senderID, tmpDir, teardown := setupTestBatchEnvironment(t)
	defer teardown()
	ctx := context.Background()

	proc := NewBatchProcessor(db, tmpDir)

	// Recipients with 2 valid, 1 duplicate, 1 invalid
	recipients := []map[string]any{
		{"phone": "01712345678"},             // Valid 1
		{"phone": "+8801712345678"},           // Duplicate of Valid 1
		{"phone": "01898765432"},             // Valid 2
		{"phone": "12345"},                   // Invalid number
	}

	configJSON, _ := json.Marshal(map[string]any{
		"recipients": recipients,
		"dedup":      true,
	})

	bRes, err := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, sender_id, body, source_type, source_config, status, created_at, updated_at)
		VALUES (?, ?, 'Test batch body', 'manual', ?, 'processing', NOW(), NOW())
	`, userID, senderID, string(configJSON))
	if err != nil {
		t.Fatalf("Failed to insert batch: %v", err)
	}
	batchID, _ := bRes.LastInsertId()

	err = proc.ProcessBatch(ctx, batchID)
	if err != nil {
		t.Fatalf("ProcessBatch returned error: %v", err)
	}

	// Verify batch status and counts
	var status string
	var validCount, invalidCount int
	var failedReportPath sql.NullString
	err = db.QueryRowContext(ctx, "SELECT status, valid_count, invalid_count, failed_report_path FROM batches WHERE id = ?", batchID).
		Scan(&status, &validCount, &invalidCount, &failedReportPath)
	if err != nil {
		t.Fatalf("Failed to scan batch: %v", err)
	}

	if status != "completed" {
		t.Errorf("Expected batch status 'completed', got '%s'", status)
	}
	if validCount != 2 {
		t.Errorf("Expected valid_count=2, got %d", validCount)
	}
	if invalidCount != 2 { // 1 duplicate + 1 invalid number
		t.Errorf("Expected invalid_count=2, got %d", invalidCount)
	}

	// Verify failed report CSV was generated
	if !failedReportPath.Valid || failedReportPath.String == "" {
		t.Fatalf("Expected failed_report_path to be generated")
	}

	fullReportPath := filepath.Join(tmpDir, failedReportPath.String)
	f, err := os.Open(fullReportPath)
	if err != nil {
		t.Fatalf("Failed to open generated failed report CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) != 3 { // Header + 2 failed records
		t.Errorf("Expected 3 rows in failed report (header + 2 rows), got %d", len(records))
	}
	if records[0][0] != "Phone" || records[0][1] != "Reason" {
		t.Errorf("Expected CSV header ['Phone', 'Reason'], got %v", records[0])
	}
}

func TestBatchProcessor_CsvAndXlsxUpload(t *testing.T) {
	db, userID, senderID, tmpDir, teardown := setupTestBatchEnvironment(t)
	defer teardown()
	ctx := context.Background()

	proc := NewBatchProcessor(db, tmpDir)

	// 1. Create CSV file
	csvRelPath := "test_recipients.csv"
	csvFullPath := filepath.Join(tmpDir, csvRelPath)
	csvFile, err := os.Create(csvFullPath)
	if err != nil {
		t.Fatalf("Failed to create CSV: %v", err)
	}
	w := csv.NewWriter(csvFile)
	_ = w.Write([]string{"Mobile", "Name"})
	_ = w.Write([]string{"01711111111", "Alice"})
	_ = w.Write([]string{"01822222222", "Bob"})
	w.Flush()
	csvFile.Close()

	// Process CSV
	csvConfig, _ := json.Marshal(map[string]any{
		"file_path":    csvRelPath,
		"phone_column": "Mobile",
		"dedup":        false,
	})

	b1Res, _ := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, sender_id, body, source_type, source_config, status, created_at, updated_at)
		VALUES (?, ?, 'Hello {{Name}}', 'upload', ?, 'processing', NOW(), NOW())
	`, userID, senderID, string(csvConfig))
	b1ID, _ := b1Res.LastInsertId()

	err = proc.ProcessBatch(ctx, b1ID)
	if err != nil {
		t.Fatalf("ProcessBatch CSV failed: %v", err)
	}

	var b1Status string
	var b1Valid int
	_ = db.QueryRowContext(ctx, "SELECT status, valid_count FROM batches WHERE id = ?", b1ID).Scan(&b1Status, &b1Valid)
	if b1Status != "completed" || b1Valid != 2 {
		t.Errorf("CSV batch assertion failed: status=%s, valid_count=%d", b1Status, b1Valid)
	}

	// 2. Create XLSX file
	xlsxRelPath := "test_recipients.xlsx"
	xlsxFullPath := filepath.Join(tmpDir, xlsxRelPath)
	xf := excelize.NewFile()
	sheet := xf.GetSheetName(0)
	_ = xf.SetCellValue(sheet, "A1", "Phone")
	_ = xf.SetCellValue(sheet, "B1", "Name")
	_ = xf.SetCellValue(sheet, "A2", "01933333333")
	_ = xf.SetCellValue(sheet, "B2", "Charlie")
	_ = xf.SetCellValue(sheet, "A3", "01644444444")
	_ = xf.SetCellValue(sheet, "B3", "David")
	if err := xf.SaveAs(xlsxFullPath); err != nil {
		t.Fatalf("Failed to save XLSX: %v", err)
	}

	// Process XLSX
	xlsxConfig, _ := json.Marshal(map[string]any{
		"file_path":    xlsxRelPath,
		"phone_column": "Phone",
		"dedup":        false,
	})

	b2Res, _ := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, sender_id, body, source_type, source_config, status, created_at, updated_at)
		VALUES (?, ?, 'Hello {{Name}}', 'upload', ?, 'processing', NOW(), NOW())
	`, userID, senderID, string(xlsxConfig))
	b2ID, _ := b2Res.LastInsertId()

	err = proc.ProcessBatch(ctx, b2ID)
	if err != nil {
		t.Fatalf("ProcessBatch XLSX failed: %v", err)
	}

	var b2Status string
	var b2Valid int
	_ = db.QueryRowContext(ctx, "SELECT status, valid_count FROM batches WHERE id = ?", b2ID).Scan(&b2Status, &b2Valid)
	if b2Status != "completed" || b2Valid != 2 {
		t.Errorf("XLSX batch assertion failed: status=%s, valid_count=%d", b2Status, b2Valid)
	}
}

func TestBatchProcessor_ApprovalFlow(t *testing.T) {
	db, userID, senderID, tmpDir, teardown := setupTestBatchEnvironment(t)
	defer teardown()
	ctx := context.Background()

	proc := NewBatchProcessor(db, tmpDir)

	configJSON, _ := json.Marshal(map[string]any{
		"recipients": []map[string]any{{"phone": "01712345678"}},
	})

	// Insert batch with approval = 'pending'
	bRes, _ := db.ExecContext(ctx, `
		INSERT INTO batches (user_id, sender_id, body, source_type, source_config, status, approval, created_at, updated_at)
		VALUES (?, ?, 'Requires approval body', 'manual', ?, 'processing', 'pending', NOW(), NOW())
	`, userID, senderID, string(configJSON))
	batchID, _ := bRes.LastInsertId()

	// Processing should skip because approval is pending
	err := proc.ProcessBatch(ctx, batchID)
	if err != nil {
		t.Fatalf("Expected nil error on skipped pending approval batch: %v", err)
	}

	var status string
	_ = db.QueryRowContext(ctx, "SELECT status FROM batches WHERE id = ?", batchID).Scan(&status)
	if status != "processing" {
		t.Errorf("Expected status to remain unchanged, got %s", status)
	}

	// Now approve it: approval = 'approved'
	_, _ = db.ExecContext(ctx, "UPDATE batches SET approval = 'approved' WHERE id = ?", batchID)
	err = proc.ProcessBatch(ctx, batchID)
	if err != nil {
		t.Fatalf("ProcessBatch failed after approval: %v", err)
	}

	_ = db.QueryRowContext(ctx, "SELECT status FROM batches WHERE id = ?", batchID).Scan(&status)
	if status != "completed" {
		t.Errorf("Expected status 'completed' after approval, got %s", status)
	}
}
