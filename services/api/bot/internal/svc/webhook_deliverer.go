package svc

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"bot/internal/dao"

	"github.com/zeromicro/go-zero/core/logx"
)

type WebhookDeliverer struct {
	eventDao *dao.EventDao
	client   *http.Client
	stopCh   chan struct{}
}

func NewWebhookDeliverer(eventDao *dao.EventDao) *WebhookDeliverer {
	return &WebhookDeliverer{
		eventDao: eventDao,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

func (w *WebhookDeliverer) Start() {
	go w.deliverLoop()
	logx.Info("Webhook deliverer started")
}

func (w *WebhookDeliverer) Stop() {
	close(w.stopCh)
	logx.Info("Webhook deliverer stopped")
}

func (w *WebhookDeliverer) deliverLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processPendingDeliveries()
		}
	}
}

func (w *WebhookDeliverer) processPendingDeliveries() {
	ctx := context.Background()
	deliveries, err := w.eventDao.QueryPendingDeliveries(ctx, 50)
	if err != nil {
		logx.Errorf("query pending deliveries: %v", err)
		return
	}

	for _, d := range deliveries {
		if d.WebhookURL == "" {
			w.eventDao.MarkFailed(ctx, d.ID, "no webhook URL configured", 0)
			continue
		}

		w.deliverEvent(d.ID, d.EventID, d.EventType, d.WebhookURL, d.WebhookSecret, d.PayloadStr, d.RetryCount, d.MaxRetries)
	}
}

func (w *WebhookDeliverer) deliverEvent(id int64, eventID, eventType, webhookURL, webhookSecret, payloadStr string, retryCount, maxRetries int) {
	ctx := context.Background()

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader([]byte(payloadStr)))
	if err != nil {
		w.eventDao.MarkFailed(ctx, id, fmt.Sprintf("create request: %v", err), 0)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Katheryne-Event-Type", eventType)
	req.Header.Set("X-Katheryne-Event-ID", eventID)
	req.Header.Set("X-Katheryne-Delivery-ID", fmt.Sprintf("%d", id))

	if webhookSecret != "" {
		sig := computeHMAC(webhookSecret, payloadStr)
		req.Header.Set("X-Katheryne-Signature", "sha256="+sig)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		backoff := computeBackoff(retryCount+1, maxRetries)
		if backoff > 0 {
			w.eventDao.ScheduleRetry(ctx, id, retryCount+1, backoff)
		} else {
			w.eventDao.MarkFailed(ctx, id, err.Error(), 0)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.eventDao.MarkDelivered(ctx, id, resp.StatusCode)
	} else if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		backoff := computeBackoff(retryCount+1, maxRetries)
		if backoff > 0 {
			w.eventDao.SetDeliveryError(ctx, id, retryCount+1, time.Now().Add(backoff), errMsg, resp.StatusCode)
		} else {
			w.eventDao.MarkFailed(ctx, id, errMsg, resp.StatusCode)
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		w.eventDao.MarkFailed(ctx, id, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), resp.StatusCode)
	}
}

func computeBackoff(attempt, maxRetries int) time.Duration {
	if attempt >= maxRetries {
		return 0
	}
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	if backoff > 60*time.Second {
		backoff = 60 * time.Second
	}
	return backoff
}

func computeHMAC(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (w *WebhookDeliverer) RecordEvent(botID, convID int64, eventType, eventID, deliveryMethod string, payload interface{}) error {
	return w.eventDao.RecordEvent(context.Background(), botID, convID, eventType, eventID, deliveryMethod, payload)
}