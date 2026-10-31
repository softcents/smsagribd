package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"smswhiteinfo/customer-api/internal/billing"
	"smswhiteinfo/customer-api/internal/crypto"
	"smswhiteinfo/customer-api/internal/sms"
	"smswhiteinfo/customer-api/internal/sms/drivers"
)

type GatewayInfo struct {
	ID                int64
	Name              string
	Driver            string
	CredentialsRaw    string
	RequestsPerSecond int
	IsEnabled         bool
}

type SendWorkerPool struct {
	db          *sql.DB
	encrypter   *crypto.LaravelEncrypter
	limiters    map[int64]*rate.Limiter
	limiterMu   sync.RWMutex
	workerCount int
	stopCh      chan struct{}
	wg          sync.WaitGroup
	inFlight    int64
	inFlightMu  sync.RWMutex
	lastLoop    time.Time
}

func NewSendWorkerPool(db *sql.DB, appKey string, workerCount int) (*SendWorkerPool, error) {
	enc, err := crypto.NewLaravelEncrypter(appKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init LaravelEncrypter in worker: %w", err)
	}

	if workerCount <= 0 {
		workerCount = 4
	}

	return &SendWorkerPool{
		db:          db,
		encrypter:   enc,
		limiters:    make(map[int64]*rate.Limiter),
		workerCount: workerCount,
		stopCh:      make(chan struct{}),
	}, nil
}

func (p *SendWorkerPool) Start(ctx context.Context) {
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("WORKER_ENABLED"))) == "true"
	if !enabled {
		log.Println("[SEND WORKER] WORKER_ENABLED is false; worker pool will remain idle.")
		return
	}

	log.Printf("[SEND WORKER] Starting %d background sending workers...", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.workerLoop(ctx, i+1)
	}
}

func (p *SendWorkerPool) Stop() {
	close(p.stopCh)

	// Wait up to 30 seconds for in-flight jobs
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[SEND WORKER] All workers stopped cleanly.")
	case <-time.After(30 * time.Second):
		log.Println("[SEND WORKER] Timed out waiting for workers after 30s. Exiting.")
	}
}

func (p *SendWorkerPool) Status() (inFlight int64, lastLoop time.Time) {
	p.inFlightMu.RLock()
	defer p.inFlightMu.RUnlock()
	return p.inFlight, p.lastLoop
}

func (p *SendWorkerPool) getLimiter(gatewayID int64, rps int) *rate.Limiter {
	p.limiterMu.Lock()
	defer p.limiterMu.Unlock()

	limiter, exists := p.limiters[gatewayID]
	if !exists {
		limit := rate.Limit(rps)
		if limit < 1 {
			limit = 1
		}
		limiter = rate.NewLimiter(limit, int(limit))
		p.limiters[gatewayID] = limiter
	}
	return limiter
}

func (p *SendWorkerPool) workerLoop(ctx context.Context, id int) {
	defer p.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.inFlightMu.Lock()
			p.lastLoop = time.Now()
			p.inFlightMu.Unlock()

			p.processNextBatch(ctx)
		}
	}
}

func (p *SendWorkerPool) processNextBatch(ctx context.Context) {
	// Determine max age threshold from env
	maxAge := 30 // minutes
	if v := os.Getenv("WORKER_MAX_AGE_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxAge = n
		}
	}

	// 1. Candidate selection (skip orphaned ancient pending messages older than maxAge)
	rows, err := p.db.QueryContext(ctx, `
		SELECT id FROM messages 
		WHERE status = 'pending' 
		  AND created_at >= DATE_SUB(NOW(), INTERVAL ? MINUTE)
		ORDER BY id ASC 
		LIMIT 50
	`, maxAge)
	if err != nil {
		return
	}
	defer rows.Close()

	var candidateIDs []int64
	for rows.Next() {
		var mid int64
		if err := rows.Scan(&mid); err == nil {
			candidateIDs = append(candidateIDs, mid)
		}
	}

	if len(candidateIDs) == 0 {
		return
	}

	// 2. Atomic claim
	inClause := strings.Repeat("?,", len(candidateIDs))
	inClause = strings.TrimSuffix(inClause, ",")

	args := make([]any, len(candidateIDs))
	for i, mid := range candidateIDs {
		args[i] = mid
	}

	res, err := p.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE messages 
		SET status = 'sending', updated_at = NOW() 
		WHERE id IN (%s) AND status = 'pending'
	`, inClause), args...)
	if err != nil {
		return
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return
	}

	// Fetch claimed messages that we successfully set to 'sending'
	q := fmt.Sprintf(`
		SELECT m.id, m.user_id, m.`+"`to`"+`, m.body, m.is_unicode, m.cost, m.gateway_id, m.sender_id,
		       s.sender_id, s.type, s.gateway_id,
		       COALESCE(g.id, sg.id), COALESCE(g.driver, sg.driver), COALESCE(g.credentials, sg.credentials), COALESCE(g.requests_per_second, sg.requests_per_second, 10), COALESCE(g.is_enabled, sg.is_enabled, 0)
		FROM messages m
		LEFT JOIN senders s ON s.id = m.sender_id
		LEFT JOIN gateways g ON g.id = m.gateway_id
		LEFT JOIN gateways sg ON sg.id = s.gateway_id
		WHERE m.id IN (%s) AND m.status = 'sending'
	`, inClause)

	claimedRows, err := p.db.QueryContext(ctx, q, args...)
	if err != nil {
		return
	}
	defer claimedRows.Close()

	for claimedRows.Next() {
		var mid, uid int64
		var to, body string
		var isUnicode bool
		var cost float64
		var mGwID, sID sql.NullInt64
		var senderLabel, senderType sql.NullString
		var sGwID sql.NullInt64
		var gwID sql.NullInt64
		var gwDriver, gwCreds sql.NullString
		var gwRps sql.NullInt64
		var gwEnabled sql.NullBool

		err := claimedRows.Scan(
			&mid, &uid, &to, &body, &isUnicode, &cost, &mGwID, &sID,
			&senderLabel, &senderType, &sGwID,
			&gwID, &gwDriver, &gwCreds, &gwRps, &gwEnabled,
		)
		if err != nil {
			continue
		}

		p.inFlightMu.Lock()
		p.inFlight++
		p.inFlightMu.Unlock()

		go func(mid int64, to, body string, isUnicode bool, sLabel string, gwID int64, gwDriver, gwCreds string, gwRps int, gwEnabled bool) {
			defer func() {
				p.inFlightMu.Lock()
				p.inFlight--
				p.inFlightMu.Unlock()
			}()

			p.sendSingleMessage(ctx, mid, to, body, isUnicode, sLabel, gwID, gwDriver, gwCreds, gwRps, gwEnabled)
		}(mid, to, body, isUnicode, senderLabel.String, gwID.Int64, gwDriver.String, gwCreds.String, int(gwRps.Int64), gwEnabled.Bool)
	}
}

func (p *SendWorkerPool) sendSingleMessage(ctx context.Context, mid int64, to, body string, isUnicode bool, senderLabel string, gwID int64, gwDriver, gwCreds string, gwRps int, gwEnabled bool) {
	// Guard: check if gateway exists and is enabled
	if gwID == 0 || !gwEnabled || gwDriver == "" || gwCreds == "" {
		_, _ = billing.RefundMessageFailed(ctx, p.db, mid, "sending", "No enabled gateway resolved for message")
		return
	}

	// Decrypt credentials
	creds, err := p.encrypter.DecryptArray(gwCreds)
	if err != nil {
		_, _ = billing.RefundMessageFailed(ctx, p.db, mid, "sending", fmt.Sprintf("Failed to decrypt gateway credentials: %v", err))
		return
	}

	// Create driver
	driver, err := drivers.CreateDriver(gwDriver, creds)
	if err != nil {
		_, _ = billing.RefundMessageFailed(ctx, p.db, mid, "sending", fmt.Sprintf("Failed to instantiate driver: %v", err))
		return
	}

	// Rate limit check
	limiter := p.getLimiter(gwID, gwRps)
	_ = limiter.Wait(ctx)

	// Send request
	req := sms.SmsRequest{
		To:        to,
		Body:      body,
		From:      senderLabel,
		IsUnicode: isUnicode,
	}

	resp, err := driver.Send(ctx, req)
	if err != nil || !resp.Successful {
		failMsg := "Gateway rejected message"
		if err != nil {
			failMsg = err.Error()
		} else if resp.ErrorMessage != "" {
			failMsg = resp.ErrorMessage
		}

		_, _ = billing.RefundMessageFailed(ctx, p.db, mid, "sending", failMsg)
		return
	}

	// Success: transition to 'sent'
	now := time.Now()
	nextPoll := now.Add(1 * time.Minute)

	gwMsgID := ""
	if resp.GatewayMessageId != nil {
		gwMsgID = *resp.GatewayMessageId
	}

	_, _ = p.db.ExecContext(ctx, `
		UPDATE messages 
		SET gateway_id = ?, status = 'sent', sent_at = ?, gateway_message_id = ?, poll_attempt = 1, next_poll_at = ?, updated_at = ?
		WHERE id = ? AND status = 'sending'
	`, gwID, now, gwMsgID, nextPoll, now, mid)
}

// ProcessNextBatchForTest exposes processNextBatch for integration tests.
// Do NOT call from production code paths.
func (p *SendWorkerPool) ProcessNextBatchForTest(ctx context.Context) {
	p.processNextBatch(ctx)
}

func NormalizeBangladeshMsisdn(value string) string {
	digits := regexp.MustCompile(`\D+`).ReplaceAllString(value, "")
	re := regexp.MustCompile(`^(?:880|0)?(1[3-9]\d{8})$`)
	m := re.FindStringSubmatch(digits)
	if len(m) == 2 {
		return "880" + m[1]
	}
	return ""
}
