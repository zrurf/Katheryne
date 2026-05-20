package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bot/internal/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventDao struct {
	db *pgxpool.Pool
}

func NewEventDao(db *pgxpool.Pool) *EventDao {
	return &EventDao{db: db}
}

type PendingDelivery struct {
	ID            int64
	EventID       string
	EventType     string
	ConvID        int64
	PayloadStr    string
	RetryCount    int
	MaxRetries    int
	BotID         int64
	WebhookURL    string
	WebhookSecret string
}

func (d *EventDao) QueryPendingDeliveries(ctx context.Context, limit int) ([]PendingDelivery, error) {
	rows, err := d.db.Query(ctx,
		`SELECT ed.id, ed.event_id, ed.event_type, ed.conv_id,
		        ed.payload::text, ed.retry_count, ed.max_retries,
		        b.bot_id, b.webhook_url, b.webhook_secret
		 FROM bot_event_delivery ed
		 JOIN bot b ON ed.bot_id = b.bot_id
		 WHERE ed.delivery_method = 'webhook'
		   AND ed.status IN ('PENDING', 'FAILED')
		   AND ed.retry_count < ed.max_retries
		   AND (ed.next_retry_at IS NULL OR ed.next_retry_at <= NOW())
		 ORDER BY ed.created_at ASC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []PendingDelivery
	for rows.Next() {
		var d PendingDelivery
		if err := rows.Scan(&d.ID, &d.EventID, &d.EventType, &d.ConvID,
			&d.PayloadStr, &d.RetryCount, &d.MaxRetries,
			&d.BotID, &d.WebhookURL, &d.WebhookSecret); err != nil {
			continue
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, nil
}

func (d *EventDao) RecordEvent(ctx context.Context, botID, convID int64, eventType, eventID, deliveryMethod string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	nextRetry := time.Now().Add(1 * time.Second)

	_, err = d.db.Exec(ctx,
		`INSERT INTO bot_event_delivery (bot_id, conv_id, event_type, event_id, payload, delivery_method, status, next_retry_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'PENDING', $7)
		 ON CONFLICT (bot_id, event_id) DO NOTHING`,
		botID, convID, eventType, eventID, payloadBytes, deliveryMethod, nextRetry)

	return err
}

func (d *EventDao) MarkDelivered(ctx context.Context, id int64, responseCode int) {
	d.db.Exec(ctx,
		`UPDATE bot_event_delivery SET status = 'DELIVERED', response_code = $1, delivered_at = NOW() WHERE id = $2`,
		responseCode, id)
}

func (d *EventDao) MarkFailed(ctx context.Context, id int64, errMsg string, responseCode int) {
	d.db.Exec(ctx,
		`UPDATE bot_event_delivery SET status = 'FAILED', last_error = $1, response_code = $2 WHERE id = $3`,
		errMsg, responseCode, id)
}

func (d *EventDao) ScheduleRetry(ctx context.Context, id int64, retryCount int, backoff time.Duration) {
	d.db.Exec(ctx,
		`UPDATE bot_event_delivery SET retry_count = $1, next_retry_at = $2, status = 'FAILED' WHERE id = $3`,
		retryCount, time.Now().Add(backoff), id)
}

func (d *EventDao) SetDeliveryError(ctx context.Context, id int64, retryCount int, nextRetry time.Time, errMsg string, responseCode int) {
	d.db.Exec(ctx,
		`UPDATE bot_event_delivery SET retry_count = $1, next_retry_at = $2, last_error = $3, status = 'FAILED', response_code = $4 WHERE id = $5`,
		retryCount, nextRetry, errMsg, responseCode, id)
}

func (d *EventDao) RetryEvent(ctx context.Context, eventID string, uid int64) error {
	var botID int64
	check := d.db.QueryRow(ctx,
		`SELECT ed.bot_id FROM bot_event_delivery ed
		 JOIN bot b ON ed.bot_id = b.bot_id
		 WHERE ed.event_id = $1 AND b.owner_uid = $2`, eventID, uid)
	if err := check.Scan(&botID); err != nil {
		return fmt.Errorf("event delivery not found or not authorized")
	}

	_, err := d.db.Exec(ctx,
		`UPDATE bot_event_delivery SET status = 'PENDING', retry_count = retry_count + 1
		 WHERE event_id = $1`, eventID)
	return err
}

func (d *EventDao) ListEventDeliveries(ctx context.Context, botID int64, convID int64, eventType, status string, page, size int) ([]types.EventDeliveryItem, int64, error) {
	var total int64
	d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bot_event_delivery WHERE bot_id = $1`, botID).Scan(&total)

	offset := (page - 1) * size

	query := `SELECT event_id, event_type, conv_id, delivery_method, status,
	                  retry_count, COALESCE(response_code, 0), COALESCE(last_error, ''),
	                  COALESCE(EXTRACT(EPOCH FROM delivered_at)::bigint, 0),
	                  EXTRACT(EPOCH FROM created_at)::bigint
	           FROM bot_event_delivery WHERE bot_id = $1`
	args := []interface{}{botID}
	argIdx := 2

	if convID > 0 {
		query += fmt.Sprintf(" AND conv_id = $%d", argIdx)
		args = append(args, convID)
		argIdx++
	}
	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, eventType)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, size, offset)

	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []types.EventDeliveryItem
	for rows.Next() {
		var item types.EventDeliveryItem
		if err := rows.Scan(&item.EventID, &item.EventType, &item.ConvID,
			&item.DeliveryMethod, &item.Status, &item.RetryCount,
			&item.ResponseCode, &item.LastError, &item.DeliveredAt, &item.CreatedAt); err != nil {
			continue
		}
		list = append(list, item)
	}

	return list, total, nil
}
