package dlr

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"smswhiteinfo/customer-api/internal/billing"
	"smswhiteinfo/customer-api/internal/crypto"
	"smswhiteinfo/customer-api/internal/scheduler"
	"smswhiteinfo/customer-api/internal/sms"
	"smswhiteinfo/customer-api/internal/sms/drivers"
)

const (
	// Matches PollMessageDeliveryJob.php line 54: public const int MAX_ATTEMPTS = 5;
	MaxPollAttempts = 5
	MaxAgeHours     = 48
)

type DeliveryPoller struct {
	db        *sql.DB
	encrypter *crypto.LaravelEncrypter
	stopCh    chan struct{}
}

func NewDeliveryPoller(db *sql.DB, appKey string) (*DeliveryPoller, error) {
	enc, err := crypto.NewLaravelEncrypter(appKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init encrypter in DLR poller: %w", err)
	}

	return &DeliveryPoller{
		db:        db,
		encrypter: enc,
		stopCh:    make(chan struct{}),
	}, nil
}

func (p *DeliveryPoller) Start(ctx context.Context) {
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("SCHEDULER_ENABLED"))) == "true"
	if !enabled {
		log.Println("[DLR POLLER] SCHEDULER_ENABLED is false; DLR polling loop will remain idle.")
		return
	}

	log.Println("[DLR POLLER] Starting 1-minute DLR polling loop...")

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-p.stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.runTick(ctx)
			}
		}
	}()
}

func (p *DeliveryPoller) Stop() {
	close(p.stopCh)
}

func (p *DeliveryPoller) runTick(ctx context.Context) {
	// Acquire per-loop lock on dedicated connection
	lock, err := scheduler.TryAcquireLock(ctx, p.db, "sms_whiteinfo:poll_delivery")
	if err != nil || lock == nil {
		return // Lock held by another instance
	}
	defer lock.Release(ctx)

	// 1. Poll due messages
	p.pollDueMessages(ctx)

	// 2. 5-Minute fallback (Correction C2): Sent for > 5 min becomes Delivered
	p.runFiveMinuteFallback(ctx)
}

// RunTickForTest exposes runTick for integration testing
func (p *DeliveryPoller) RunTickForTest(ctx context.Context) {
	p.runTick(ctx)
}

func (p *DeliveryPoller) pollDueMessages(ctx context.Context) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT m.id, m.`+"`to`"+`, m.gateway_message_id, COALESCE(m.poll_attempt, 1),
		       COALESCE(g.id, sg.id), COALESCE(g.driver, sg.driver), COALESCE(g.credentials, sg.credentials)
		FROM messages m
		LEFT JOIN senders s ON s.id = m.sender_id
		LEFT JOIN gateways g ON g.id = m.gateway_id
		LEFT JOIN gateways sg ON sg.id = s.gateway_id
		WHERE m.status = 'sent'
		  AND m.gateway_message_id IS NOT NULL
		  AND m.next_poll_at IS NOT NULL
		  AND m.next_poll_at <= NOW()
		  AND m.sent_at >= DATE_SUB(NOW(), INTERVAL 48 HOUR)
		ORDER BY m.next_poll_at ASC
		LIMIT 50
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	type dueItem struct {
		id          int64
		to          string
		gwMsgID     string
		pollAttempt int
		gwID        int64
		gwDriver    string
		gwCreds     string
	}

	var items []dueItem
	for rows.Next() {
		var it dueItem
		var gwID sql.NullInt64
		var gwDriver, gwCreds sql.NullString
		if err := rows.Scan(&it.id, &it.to, &it.gwMsgID, &it.pollAttempt, &gwID, &gwDriver, &gwCreds); err == nil {
			it.gwID = gwID.Int64
			it.gwDriver = gwDriver.String
			it.gwCreds = gwCreds.String
			items = append(items, it)
		}
	}

	for _, it := range items {
		if it.gwID == 0 || it.gwDriver == "" || it.gwCreds == "" {
			// Stop polling if gateway cannot be resolved
			_, _ = p.db.ExecContext(ctx, "UPDATE messages SET poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() WHERE id = ?", it.id)
			continue
		}

		creds, cErr := p.encrypter.DecryptArray(it.gwCreds)
		if cErr != nil {
			continue
		}

		driver, dErr := drivers.CreateDriver(it.gwDriver, creds)
		if dErr != nil {
			continue
		}

		dlrDriver, ok := driver.(sms.SupportsDeliveryStatus)
		if !ok {
			// Gateway driver doesn't support query DLR, stop polling
			_, _ = p.db.ExecContext(ctx, "UPDATE messages SET poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() WHERE id = ?", it.id)
			continue
		}

		resp, err := dlrDriver.FetchDeliveryStatus(ctx, it.gwMsgID, it.to)
		if err != nil {
			// Reschedule poll
			p.scheduleNextPoll(ctx, it.id, it.pollAttempt)
			continue
		}

		now := time.Now()
		switch resp.Outcome {
		case sms.OutcomeDelivered:
			deliveredAt := now
			if resp.DeliveredAt != nil {
				deliveredAt = *resp.DeliveredAt
			}
			_, _ = p.db.ExecContext(ctx, `
				UPDATE messages 
				SET status = 'delivered', delivered_at = ?, delivery_status = ?, poll_attempt = NULL, next_poll_at = NULL, updated_at = ?
				WHERE id = ? AND status = 'sent'
			`, deliveredAt, resp.RawStatus, now, it.id)

		case sms.OutcomeFailed:
			reason := resp.ErrorMessage
			if reason == "" {
				reason = resp.RawStatus
			}
			_, _ = billing.RefundMessageFailed(ctx, p.db, it.id, "sent", reason)

		case sms.OutcomeInTransit:
			p.scheduleNextPoll(ctx, it.id, it.pollAttempt)
		}
	}
}

func (p *DeliveryPoller) scheduleNextPoll(ctx context.Context, messageID int64, currentAttempt int) {
	if currentAttempt >= MaxPollAttempts {
		// Stop polling and retain 'sent' status
		_, _ = p.db.ExecContext(ctx, "UPDATE messages SET poll_attempt = NULL, next_poll_at = NULL, updated_at = NOW() WHERE id = ?", messageID)
		return
	}

	nextAttempt := currentAttempt + 1
	var delaySec int
	// Matches PollMessageDeliveryJob.php lines 206-214:
	// 1 => 60 (1 min initial), 2 => 120 (+2 min), 3 => 120 (+2 min), default => 120
	switch nextAttempt {
	case 1:
		delaySec = 60 // 1 min (initial)
	case 2:
		delaySec = 120 // +2 min
	case 3:
		delaySec = 120 // +2 min
	default:
		delaySec = 120
	}

	nextPoll := time.Now().Add(time.Duration(delaySec) * time.Second)
	_, _ = p.db.ExecContext(ctx, `
		UPDATE messages 
		SET poll_attempt = ?, next_poll_at = ?, updated_at = NOW() 
		WHERE id = ?
	`, nextAttempt, nextPoll, messageID)
}

func (p *DeliveryPoller) runFiveMinuteFallback(ctx context.Context) {
	// Fallback applies EXCLUSIVELY to status = 'sent' with sent_at <= now() - 5 minutes
	// Matches Laravel: Message::where('status', 'sent')->where('sent_at', '<=', now()->subMinutes(5))
	threshold := time.Now().Add(-5 * time.Minute)
	rows, err := p.db.QueryContext(ctx, `
		SELECT id FROM messages
		WHERE status = 'sent' 
		  AND sent_at <= ?
		ORDER BY id ASC
		LIMIT 2000
	`, threshold)
	if err != nil {
		return
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return
	}

	// Update in chunks of 200
	const chunkSize = 200
	now := time.Now()

	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[i:end]

		inClause := strings.Repeat("?,", len(chunk))
		inClause = strings.TrimSuffix(inClause, ",")

		args := make([]any, len(chunk)+2)
		args[0] = now
		args[1] = now
		for j, id := range chunk {
			args[j+2] = id
		}

		q := fmt.Sprintf(`
			UPDATE messages 
			SET status = 'delivered', delivered_at = ?, poll_attempt = NULL, next_poll_at = NULL, updated_at = ?
			WHERE id IN (%s) AND status = 'sent'
		`, inClause)

		_, _ = p.db.ExecContext(ctx, q, args...)
	}
}
