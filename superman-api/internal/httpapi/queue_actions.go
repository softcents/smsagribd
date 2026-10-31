package httpapi

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
)

func (server *Server) queueAction(context fiber.Ctx) error {
	action := context.Params("action")
	admin := superman(context)

	var affected int64
	var err error
	var message string

	switch action {
	case "pause":
		affected, err = server.setQueuePaused(true)
		message = "All SMS dispatch queues paused."
	case "resume":
		affected, err = server.setQueuePaused(false)
		message = "All SMS dispatch queues resumed."
	case "retry-failed":
		result, updateErr := server.db.Exec(`
			UPDATE messages
			SET status = 'pending', failed_reason = NULL, updated_at = NOW()
			WHERE status IN ('failed', 'dlr_failed')
				AND updated_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
			LIMIT 500
		`)
		err = updateErr
		if result != nil {
			affected, _ = result.RowsAffected()
		}
		message = fmt.Sprintf("Queued %d failed messages for retry.", affected)
	case "requeue-stuck":
		result, updateErr := server.db.Exec(`
			UPDATE messages
			SET status = 'pending', updated_at = NOW()
			WHERE status = 'sending' AND updated_at <= DATE_SUB(NOW(), INTERVAL 2 MINUTE)
		`)
		err = updateErr
		if result != nil {
			affected, _ = result.RowsAffected()
		}
		message = fmt.Sprintf("Requeued %d stuck messages.", affected)
	case "purge":
		affected, err = server.purgeQueue()
		message = fmt.Sprintf("Purged %d queue items.", affected)
	default:
		return context.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Unknown queue action."})
	}

	if err != nil {
		return databaseError(context, err)
	}

	return context.JSON(fiber.Map{
		"message": message,
		"data": fiber.Map{
			"affected":     affected,
			"performed_by": admin.ID,
		},
	})
}

func (server *Server) setQueuePaused(paused bool) (int64, error) {
	key := server.config.CachePrefix + "sms_queue_paused_by_admin"
	if !paused {
		result, err := server.db.Exec("DELETE FROM cache WHERE `key` = ?", key)
		if err != nil {
			return 0, err
		}

		return result.RowsAffected()
	}

	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	result, err := server.db.Exec(`
		INSERT INTO cache (`+"`key`"+`, value, expiration)
		VALUES (?, 'b:1;', ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value), expiration = VALUES(expiration)
	`, key, expiresAt)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (server *Server) purgeQueue() (int64, error) {
	transaction, err := server.db.Begin()
	if err != nil {
		return 0, err
	}
	defer transaction.Rollback()

	messages, err := transaction.Exec(`
		UPDATE messages
		SET status = 'failed', failed_reason = 'Queue purged by Administrator', updated_at = NOW()
		WHERE status IN ('pending', 'sending')
	`)
	if err != nil {
		return 0, err
	}

	batches, err := transaction.Exec(`
		UPDATE batches
		SET status = 'failed', updated_at = NOW()
		WHERE status IN ('pending', 'processing')
	`)
	if err != nil {
		return 0, err
	}

	if err := transaction.Commit(); err != nil {
		return 0, err
	}

	messageCount, _ := messages.RowsAffected()
	batchCount, _ := batches.RowsAffected()

	return messageCount + batchCount, nil
}

// queueMonitor monitors RabbitMQ dispatch queues.
func (server *Server) queueMonitor(context fiber.Ctx) error {
	var pending, sending, sent, failed, stale int
	var oldestAgeSec sql.NullInt64

	_ = server.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending_cnt,
			COALESCE(SUM(CASE WHEN status = 'sending' THEN 1 ELSE 0 END), 0) as sending_cnt,
			COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0) as sent_cnt,
			COALESCE(SUM(CASE WHEN status IN ('failed', 'dlr_failed') THEN 1 ELSE 0 END), 0) as failed_cnt,
			COALESCE(SUM(CASE WHEN status IN ('pending', 'sending') AND updated_at <= DATE_SUB(NOW(), INTERVAL 8 MINUTE) THEN 1 ELSE 0 END), 0) as stale_cnt,
			TIMESTAMPDIFF(SECOND, MIN(CASE WHEN status = 'pending' THEN created_at END), NOW()) as oldest_pending_sec
		FROM messages
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
	`).Scan(&pending, &sending, &sent, &failed, &stale, &oldestAgeSec)

	workerStatus := "idle"
	if pending > 0 || sending > 0 {
		workerStatus = "active"
	}

	oldest := int64(0)
	if oldestAgeSec.Valid && oldestAgeSec.Int64 > 0 {
		oldest = oldestAgeSec.Int64
	}

	return context.JSON(fiber.Map{
		"data": fiber.Map{
			"pending_count":              pending,
			"sending_count":              sending,
			"sent_count":                 sent,
			"failed_count":               failed,
			"stale_count":                stale,
			"oldest_pending_age_seconds": oldest,
			"worker_status":              workerStatus,
			"timestamp":                  time.Now().Format(time.RFC3339),
		},
	})
}
