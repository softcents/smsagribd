package batch

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"smswhiteinfo/customer-api/internal/billing"
)

var bdPhoneRegex = regexp.MustCompile(`^(?:\+?88)?0?(1[3-9]\d{8})$`)

func FormatByRegex(raw string) (string, bool) {
	digits := regexp.MustCompile(`\D+`).ReplaceAllString(raw, "")
	m := bdPhoneRegex.FindStringSubmatch(digits)
	if len(m) == 2 {
		return "+880" + m[1], true
	}
	return "", false
}

type RecipientItem struct {
	Cursor int64
	Phone  string
	Values map[string]string
	Raw    map[string]any
}

type MessagePublisher interface {
	PublishMessage(ctx context.Context, msgID int64) error
}

type BatchProcessor struct {
	db          *sql.DB
	storageRoot string
	publisher   MessagePublisher
}

func NewBatchProcessor(db *sql.DB, storageRoot string) *BatchProcessor {
	if storageRoot == "" {
		storageRoot = "./storage/app/private"
	}
	return &BatchProcessor{
		db:          db,
		storageRoot: storageRoot,
	}
}

func (p *BatchProcessor) SetPublisher(pub MessagePublisher) {
	p.publisher = pub
}

// ProcessBatch runs the full parsing, validation, chunking, pricing, and message insertion for a batch.
func (p *BatchProcessor) ProcessBatch(ctx context.Context, batchID int64) error {
	var userID int64
	var senderID sql.NullInt64
	var body string
	var sourceType string
	var sourceIP sql.NullString
	var sourceConfigRaw []byte
	var status string
	var approval sql.NullString
	var lastCursor sql.NullInt64
	var validCount, invalidCount int

	err := p.db.QueryRowContext(ctx, `
		SELECT user_id, sender_id, body, source_type, source_ip, source_config, status, approval, last_processed_cursor, valid_count, invalid_count
		FROM batches
		WHERE id = ?
	`, batchID).Scan(&userID, &senderID, &body, &sourceType, &sourceIP, &sourceConfigRaw, &status, &approval, &lastCursor, &validCount, &invalidCount)
	if err != nil {
		return err
	}

	if status != "processing" {
		return nil // Not in processing state
	}

	if approval.Valid && approval.String == "pending" {
		return nil // Requires admin approval
	}

	if !senderID.Valid {
		_, _ = p.db.ExecContext(ctx, "UPDATE batches SET status = 'failed', updated_at = NOW() WHERE id = ?", batchID)
		return fmt.Errorf("batch %d has no sender_id", batchID)
	}

	// Fetch sender details
	var senderLabel string
	var senderType string
	var gatewayID sql.NullInt64
	err = p.db.QueryRowContext(ctx, `
		SELECT s.sender_id, s.type, s.gateway_id
		FROM senders s
		WHERE s.id = ?
	`, senderID.Int64).Scan(&senderLabel, &senderType, &gatewayID)
	if err != nil {
		_, _ = p.db.ExecContext(ctx, "UPDATE batches SET status = 'failed', updated_at = NOW() WHERE id = ?", batchID)
		return fmt.Errorf("sender not found for batch %d: %w", batchID, err)
	}

	isMasking := strings.Contains(strings.ToLower(senderType), "masking")

	rateResolver, err := billing.ForUser(ctx, p.db, userID)
	if err != nil {
		_, _ = p.db.ExecContext(ctx, "UPDATE batches SET status = 'failed', updated_at = NOW() WHERE id = ?", batchID)
		return fmt.Errorf("failed to init rate resolver: %w", err)
	}

	var sourceConfig map[string]any
	if len(sourceConfigRaw) > 0 {
		_ = json.Unmarshal(sourceConfigRaw, &sourceConfig)
	}

	// Setup failed report CSV (exact Laravel parity: batches/<id>/failed-recipients-<Ymd-His>.csv)
	relReportPath := fmt.Sprintf("batches/%d/failed-recipients-%s.csv", batchID, time.Now().Format("20060102-150405"))
	reportFilePath := filepath.Join(p.storageRoot, relReportPath)
	_ = os.MkdirAll(filepath.Dir(reportFilePath), 0755)

	reportFile, fErr := os.OpenFile(reportFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	var csvWriter *csv.Writer
	if fErr == nil {
		csvWriter = csv.NewWriter(reportFile)
		_ = csvWriter.Write([]string{"Phone", "Reason"})
		csvWriter.Flush()
	}

	defer func() {
		if reportFile != nil {
			reportFile.Close()
		}
	}()

	bodyVars := extractBodyVariables(body)

	dedup := false
	if d, ok := sourceConfig["dedup"].(bool); ok && d {
		dedup = true
	}
	seenPhones := make(map[string]bool)

	// Fetch recipients stream
	recipients, err := p.extractRecipients(ctx, sourceType, sourceConfig, lastCursor)
	if err != nil {
		_, _ = p.db.ExecContext(ctx, "UPDATE batches SET status = 'failed', updated_at = NOW() WHERE id = ?", batchID)
		return err
	}

	const chunkSize = 1000
	var buffer []map[string]any
	var cursor int64
	now := time.Now()

	recordFailure := func(phone, reason string) {
		invalidCount++
		if csvWriter != nil {
			_ = csvWriter.Write([]string{phone, reason})
			csvWriter.Flush()
		}
	}

	flushChunk := func() error {
		if len(buffer) == 0 {
			return nil
		}

		tx, tErr := p.db.BeginTx(ctx, nil)
		if tErr != nil {
			return tErr
		}
		defer tx.Rollback()

		var totalCost float64

		// Calculate rate and cost for each item
		for _, row := range buffer {
			phone := row["to"].(string)
			parts := row["parts"].(int)

			rate, rErr := rateResolver.Resolve(phone, isMasking)
			if rErr != nil {
				return rErr
			}

			cost := billing.RoundFloat(rate*float64(parts), 4)
			row["rate"] = rate
			row["cost"] = cost
			row["operator_name"] = rateResolver.OperatorNameFor(phone)
			row["tariff_name"] = rateResolver.PlanName()

			totalCost = billing.RoundFloat(totalCost+cost, 4)
		}

		// Debit user balance inside transaction
		_, dErr := billing.DebitBalance(ctx, tx, userID, totalCost, "message_batch", batchID, fmt.Sprintf("SMS batch debit (%d messages)", len(buffer)))
		if dErr != nil {
			return dErr
		}

		// Bulk insert messages
		valueStrings := make([]string, 0, len(buffer))
		valueArgs := make([]any, 0, len(buffer)*14)

		for _, row := range buffer {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())")
			valueArgs = append(valueArgs,
				userID,
				batchID,
				senderID.Int64,
				gatewayID,
				row["to"],
				row["body"],
				row["source"],
				row["source_ip"],
				row["operator_name"],
				row["tariff_name"],
				row["cost"],
				row["rate"],
				row["parts"],
				row["is_unicode"],
			)
		}

		insertStmt := fmt.Sprintf(`
			INSERT INTO messages (
				user_id, batch_id, sender_id, gateway_id, `+"`to`"+`, body, status, source, source_ip, operator_name, tariff_name, cost, rate, parts, is_unicode, created_at, updated_at
			) VALUES %s
		`, strings.Join(valueStrings, ","))

		_, iErr := tx.ExecContext(ctx, insertStmt, valueArgs...)
		if iErr != nil {
			return fmt.Errorf("failed to bulk insert messages: %w", iErr)
		}

		// Advance cursor and counts on batch
		_, uErr := tx.ExecContext(ctx, `
			UPDATE batches 
			SET valid_count = ?, invalid_count = ?, last_processed_cursor = ?, updated_at = NOW() 
			WHERE id = ?
		`, validCount, invalidCount, cursor, batchID)
		if uErr != nil {
			return uErr
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		if p.publisher != nil {
			if qRows, qErr := p.db.QueryContext(ctx, "SELECT id FROM messages WHERE batch_id = ? AND status = 'pending' ORDER BY id ASC", batchID); qErr == nil {
				for qRows.Next() {
					var mid int64
					if sErr := qRows.Scan(&mid); sErr == nil {
						_ = p.publisher.PublishMessage(ctx, mid)
					}
				}
				qRows.Close()
			}
		}

		buffer = buffer[:0]
		return nil
	}

	for _, item := range recipients {
		cursor = item.Cursor
		phoneRaw := strings.TrimSpace(item.Phone)
		if phoneRaw == "" {
			recordFailure(phoneRaw, "Phone number missing")
			continue
		}

		formatted, ok := FormatByRegex(phoneRaw)
		if !ok {
			recordFailure(phoneRaw, "Invalid phone number format")
			continue
		}

		if dedup {
			if seenPhones[formatted] {
				recordFailure(phoneRaw, "Duplicate phone number")
				continue
			}
			seenPhones[formatted] = true
		}

		// Check missing template variables
		missing := false
		for _, v := range bodyVars {
			val := strings.TrimSpace(item.Values[v])
			if val == "" {
				recordFailure(phoneRaw, fmt.Sprintf("Missing value for: %s", v))
				missing = true
				break
			}
		}
		if missing {
			continue
		}

		finalBody := body
		for _, v := range bodyVars {
			finalBody = strings.ReplaceAll(finalBody, "{"+v+"}", item.Values[v])
		}

		parts, isUni := billing.CalculateParts(finalBody)

		src := "web"
		if sourceType == "api" {
			src = "api"
		}

		buffer = append(buffer, map[string]any{
			"to":         formatted,
			"body":       finalBody,
			"parts":      parts,
			"is_unicode": isUni,
			"source":     src,
			"source_ip":  sourceIP,
			"created_at": now,
		})

		validCount++

		if len(buffer) >= chunkSize {
			if err := flushChunk(); err != nil {
				// Insufficient balance or error
				_, _ = p.db.ExecContext(ctx, `
					UPDATE batches 
					SET status = 'failed', failed_report_path = ?, updated_at = NOW() 
					WHERE id = ?
				`, reportFilePath, batchID)
				return err
			}
		}
	}

	if len(buffer) > 0 {
		if err := flushChunk(); err != nil {
			_, _ = p.db.ExecContext(ctx, `
				UPDATE batches 
				SET status = 'failed', failed_report_path = ?, updated_at = NOW() 
				WHERE id = ?
			`, relReportPath, batchID)
			return err
		}
	}

	// Mark completed
	var savedReportPath *string
	if invalidCount > 0 {
		savedReportPath = &relReportPath
	}

	_, _ = p.db.ExecContext(ctx, `
		UPDATE batches 
		SET status = 'completed', completed_at = NOW(), failed_report_path = ?, updated_at = NOW() 
		WHERE id = ?
	`, savedReportPath, batchID)

	return nil
}

func (p *BatchProcessor) extractRecipients(ctx context.Context, sourceType string, config map[string]any, lastCursor sql.NullInt64) ([]RecipientItem, error) {
	var items []RecipientItem
	startCursor := int64(0)
	if lastCursor.Valid {
		startCursor = lastCursor.Int64
	}

	switch sourceType {
	case "manual", "api":
		recRaw, ok := config["recipients"].([]any)
		if !ok {
			return nil, nil
		}
		for i, r := range recRaw {
			cur := int64(i + 1)
			if cur <= startCursor {
				continue
			}
			rm, ok := r.(map[string]any)
			if !ok {
				continue
			}
			phone := fmt.Sprintf("%v", rm["phone"])
			vals := make(map[string]string)
			if vm, ok := rm["values"].(map[string]any); ok {
				for k, v := range vm {
					vals[k] = fmt.Sprintf("%v", v)
				}
			}
			items = append(items, RecipientItem{
				Cursor: cur,
				Phone:  phone,
				Values: vals,
				Raw:    rm,
			})
		}

	case "group":
		groupIdsRaw, ok := config["group_ids"].([]any)
		if !ok || len(groupIdsRaw) == 0 {
			return nil, nil
		}
		var groupIDs []int64
		for _, g := range groupIdsRaw {
			switch v := g.(type) {
			case float64:
				groupIDs = append(groupIDs, int64(v))
			case int64:
				groupIDs = append(groupIDs, v)
			}
		}

		if len(groupIDs) == 0 {
			return nil, nil
		}

		inParams := strings.Repeat("?,", len(groupIDs))
		inParams = strings.TrimSuffix(inParams, ",")

		args := make([]any, len(groupIDs)+1)
		for i, id := range groupIDs {
			args[i] = id
		}
		args[len(groupIDs)] = startCursor

		q := fmt.Sprintf(`
			SELECT DISTINCT c.id, c.phone, COALESCE(c.name, '')
			FROM contacts c
			JOIN contact_group cg ON cg.contact_id = c.id
			WHERE cg.group_id IN (%s) AND c.deleted_at IS NULL AND c.id > ?
			ORDER BY c.id ASC
		`, inParams)

		rows, err := p.db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var cid int64
			var phone, name string
			if err := rows.Scan(&cid, &phone, &name); err == nil {
				items = append(items, RecipientItem{
					Cursor: cid,
					Phone:  phone,
					Values: map[string]string{"name": name},
					Raw:    map[string]any{"phone": phone, "name": name},
				})
			}
		}

	case "upload":
		filePath := fmt.Sprintf("%v", config["file_path"])
		if filePath == "" {
			return nil, fmt.Errorf("file_path missing in upload config")
		}

		fullPath := filePath
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(p.storageRoot, filePath)
		}

		ext := strings.ToLower(filepath.Ext(fullPath))
		if ext == ".xlsx" {
			return parseXlsxRecipients(fullPath, config, startCursor)
		}
		return parseCsvRecipients(fullPath, config, startCursor)
	}

	return items, nil
}

func parseCsvRecipients(path string, config map[string]any, startCursor int64) ([]RecipientItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open upload CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	mapping, _ := config["mapping"].(map[string]any)
	phoneCol := "phone"
	if mapping != nil && mapping["phone"] != nil {
		phoneCol = fmt.Sprintf("%v", mapping["phone"])
	}

	phoneIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), strings.TrimSpace(phoneCol)) {
			phoneIdx = i
			break
		}
	}
	if phoneIdx == -1 {
		phoneIdx = 0 // default to first column
	}

	var items []RecipientItem
	var cursor int64

	for {
		record, rErr := reader.Read()
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			continue
		}
		cursor++
		if cursor <= startCursor {
			continue
		}

		phone := ""
		if phoneIdx < len(record) {
			phone = record[phoneIdx]
		}

		vals := make(map[string]string)
		for i, h := range header {
			if i < len(record) {
				vals[strings.TrimSpace(h)] = record[i]
			}
		}

		items = append(items, RecipientItem{
			Cursor: cursor,
			Phone:  phone,
			Values: vals,
			Raw:    valsToAny(vals),
		})
	}

	return items, nil
}

func parseXlsxRecipients(path string, config map[string]any, startCursor int64) ([]RecipientItem, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in XLSX")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("failed to read rows from XLSX: %w", err)
	}

	header := rows[0]
	mapping, _ := config["mapping"].(map[string]any)
	phoneCol := "phone"
	if mapping != nil && mapping["phone"] != nil {
		phoneCol = fmt.Sprintf("%v", mapping["phone"])
	}

	phoneIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), strings.TrimSpace(phoneCol)) {
			phoneIdx = i
			break
		}
	}
	if phoneIdx == -1 {
		phoneIdx = 0
	}

	var items []RecipientItem
	for i := 1; i < len(rows); i++ {
		cursor := int64(i)
		if cursor <= startCursor {
			continue
		}
		record := rows[i]
		phone := ""
		if phoneIdx < len(record) {
			phone = record[phoneIdx]
		}

		vals := make(map[string]string)
		for j, h := range header {
			if j < len(record) {
				vals[strings.TrimSpace(h)] = record[j]
			}
		}

		items = append(items, RecipientItem{
			Cursor: cursor,
			Phone:  phone,
			Values: vals,
			Raw:    valsToAny(vals),
		})
	}

	return items, nil
}

func extractBodyVariables(body string) []string {
	re := regexp.MustCompile(`\{([a-zA-Z0-9_-]+)\}`)
	matches := re.FindAllStringSubmatch(body, -1)
	seen := make(map[string]bool)
	var vars []string
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			vars = append(vars, m[1])
		}
	}
	return vars
}

func valsToAny(vals map[string]string) map[string]any {
	res := make(map[string]any, len(vals))
	for k, v := range vals {
		res[k] = v
	}
	return res
}
