package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/batch"
	"smswhiteinfo/customer-api/internal/billing"
)

type MasterScheduler struct {
	db        *sql.DB
	processor *batch.BatchProcessor
	stopCh    chan struct{}
}

func NewMasterScheduler(db *sql.DB, processor *batch.BatchProcessor) *MasterScheduler {
	return &MasterScheduler{
		db:        db,
		processor: processor,
		stopCh:    make(chan struct{}),
	}
}

func (s *MasterScheduler) Start(ctx context.Context) {
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("SCHEDULER_ENABLED"))) == "true"
	if !enabled {
		log.Println("[SCHEDULER] SCHEDULER_ENABLED is false; scheduler loops will remain idle.")
		return
	}

	log.Println("[SCHEDULER] Starting background scheduler loops...")

	// 1. Send Scheduled Batches (every 10 seconds)
	go s.loop(ctx, 10*time.Second, "sms_whiteinfo:batches_send_scheduled", s.runSendScheduledBatches)

	// 2. Purge Stale Queue (every 1 minute)
	go s.loop(ctx, 1*time.Minute, "sms_whiteinfo:purge_stale", func(ctx context.Context) {
		s.runPurgeStale(ctx, 8)
	})

	// 3. Send Scheduled Campaigns (every 1 minute)
	go s.loop(ctx, 1*time.Minute, "sms_whiteinfo:campaigns", s.runSendScheduledCampaigns)

	// 4. Telegram Queue Alert (every 1 minute)
	go s.loop(ctx, 1*time.Minute, "sms_whiteinfo:telegram_alerts", s.runTelegramAlertCheck)
}

func (s *MasterScheduler) Stop() {
	close(s.stopCh)
}

func (s *MasterScheduler) loop(ctx context.Context, interval time.Duration, lockName string, task func(context.Context)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			lock, err := TryAcquireLock(ctx, s.db, lockName)
			if err != nil || lock == nil {
				continue // Lock held by another instance
			}

			task(ctx)
			_ = lock.Release(ctx)
		}
	}
}

func (s *MasterScheduler) runSendScheduledBatches(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, approval
		FROM batches
		WHERE status = 'scheduled'
		  AND scheduled_at IS NOT NULL
		  AND scheduled_at <= NOW()
		ORDER BY scheduled_at ASC
		LIMIT 50
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	type batchItem struct {
		id       int64
		approval sql.NullString
	}

	var dueBatches []batchItem
	for rows.Next() {
		var it batchItem
		if err := rows.Scan(&it.id, &it.approval); err == nil {
			dueBatches = append(dueBatches, it)
		}
	}

	for _, b := range dueBatches {
		targetStatus := "processing"
		if b.approval.Valid && b.approval.String == "pending" {
			targetStatus = "pending"
		}

		res, err := s.db.ExecContext(ctx, `
			UPDATE batches 
			SET status = ?, updated_at = NOW() 
			WHERE id = ? AND status = 'scheduled'
		`, targetStatus, b.id)
		if err != nil {
			continue
		}

		affected, _ := res.RowsAffected()
		if affected == 1 && targetStatus == "processing" {
			go func(bid int64) {
				_ = s.processor.ProcessBatch(context.Background(), bid)
			}(b.id)
		}
	}
}

func (s *MasterScheduler) runPurgeStale(ctx context.Context, minutes int) {
	// Guard (Item 7): If WORKER_ENABLED is false, do not purge stale queue!
	// This prevents purging messages waiting in queue during maintenance or staggered deployment.
	workerEnabled := strings.ToLower(strings.TrimSpace(os.Getenv("WORKER_ENABLED"))) == "true"
	if !workerEnabled {
		log.Println("[SCHEDULER] WARNING: WORKER_ENABLED is false; purge-stale skipped to avoid purging unconsumed queue.")
		return
	}

	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)
	// Default to true unless explicitly disabled with PURGE_STALE_REFUND=false
	purgeEnv := strings.ToLower(strings.TrimSpace(os.Getenv("PURGE_STALE_REFUND")))
	refundEnabled := purgeEnv != "false"

	if refundEnabled {
		// Refund when purging stale sending/pending messages (Item 3 & 6)
		rows, err := s.db.QueryContext(ctx, `
			SELECT id, user_id, cost 
			FROM messages 
			WHERE status IN ('pending', 'sending') AND updated_at <= ?
		`, threshold)
		if err == nil {
			type msgToRefund struct {
				id     int64
				userID int64
				cost   float64
			}
			var msgs []msgToRefund
			for rows.Next() {
				var m msgToRefund
				if err := rows.Scan(&m.id, &m.userID, &m.cost); err == nil {
					msgs = append(msgs, m)
				}
			}
			rows.Close()

			for _, m := range msgs {
				tx, err := s.db.BeginTx(ctx, nil)
				if err != nil {
					continue
				}
				res, err := tx.ExecContext(ctx, `
					UPDATE messages 
					SET status = 'failed', failed_at = NOW(), failed_reason = ?, updated_at = NOW() 
					WHERE id = ? AND status IN ('pending', 'sending')
				`, fmt.Sprintf("Auto-purged: stuck in queue for more than %d minutes without progress", minutes), m.id)
				if err != nil {
					_ = tx.Rollback()
					continue
				}
				affected, _ := res.RowsAffected()
				if affected == 1 && m.cost > 0 {
					_, _ = billing.CreditBalance(ctx, tx, m.userID, m.cost, "message", m.id, "Refund: Auto-purged stale message")
				}
				_ = tx.Commit()
			}
		}
	} else {
		// Mark failed without refunding
		_, _ = s.db.ExecContext(ctx, `
			UPDATE messages 
			SET status = 'failed', failed_at = NOW(), failed_reason = ?, updated_at = NOW() 
			WHERE status IN ('pending', 'sending') AND updated_at <= ?
		`, fmt.Sprintf("Auto-purged: stuck in queue for more than %d minutes without progress", minutes), threshold)
	}

	// Purge stale batches
	_, _ = s.db.ExecContext(ctx, `
		UPDATE batches 
		SET status = 'failed', updated_at = NOW() 
		WHERE status IN ('pending', 'processing') AND updated_at <= ?
	`, threshold)
}

func (s *MasterScheduler) runSendScheduledCampaigns(ctx context.Context) {
	// Campaigns table check (0 rows on prod)
	_, _ = s.db.ExecContext(ctx, `
		UPDATE campaigns 
		SET status = 'processing', updated_at = NOW() 
		WHERE status = 'scheduled' AND scheduled_at IS NOT NULL AND scheduled_at <= NOW()
	`)
}

func (s *MasterScheduler) runTelegramAlertCheck(ctx context.Context) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if botToken == "" || chatID == "" {
		return
	}

	var oldestAgeSec sql.NullInt64
	var pendingCount, sendingCount int
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			TIMESTAMPDIFF(SECOND, MIN(created_at), NOW()) as oldest_sec,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN status = 'sending' THEN 1 ELSE 0 END), 0) as sending
		FROM messages
		WHERE status IN ('pending', 'sending')
	`).Scan(&oldestAgeSec, &pendingCount, &sendingCount)

	if err != nil || !oldestAgeSec.Valid {
		return
	}

	// Threshold matching RabbitMQ queue wait time (60 seconds)
	if oldestAgeSec.Int64 > 60 && (pendingCount+sendingCount) > 0 {
		msg := fmt.Sprintf(
			"<b>SMS-WhiteInfo: Long Queue Wait Detected</b>\n\n"+
				"Pending Messages: <b>%d</b>\n"+
				"Sending Messages: <b>%d</b>\n"+
				"Oldest Message Wait Time: <b>%d seconds</b>",
			pendingCount, sendingCount, oldestAgeSec.Int64,
		)

		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
		data := url.Values{}
		data.Set("chat_id", chatID)
		data.Set("text", msg)
		data.Set("parse_mode", "HTML")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, pErr := client.PostForm(apiURL, data)
		if pErr == nil {
			_ = resp.Body.Close()
		}
	}
}

// RunPurgeStaleForTest exposes runPurgeStale for integration testing
func (s *MasterScheduler) RunPurgeStaleForTest(ctx context.Context, minutes int) {
	s.runPurgeStale(ctx, minutes)
}
