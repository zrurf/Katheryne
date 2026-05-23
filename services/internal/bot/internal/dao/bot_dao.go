package dao

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"bot/bot"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BotDao struct {
	db *pgxpool.Pool
}

func NewBotDao(db *pgxpool.Pool) *BotDao {
	return &BotDao{db: db}
}

func (d *BotDao) CreateBot(ctx context.Context, uid int64, req *bot.CreateBotReq) (*bot.CreateBotResp, error) {
	botID := time.Now().UnixNano()
	clientID := "bot_" + randomHex(16)
	clientSecret := randomHex(32)
	webhookSecret := randomHex(16)

	subscribeEvents := req.SubscribeEvents
	if len(subscribeEvents) == 0 {
		subscribeEvents = []string{"message.create"}
	}

	_, err := d.db.Exec(ctx,
		`INSERT INTO bot (bot_id, name, avatar, description, owner_uid, webhook_url, webhook_secret, subscribe_events, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ACTIVE')`,
		botID, req.Name, req.Avatar, req.Description, uid,
		req.WebhookUrl, webhookSecret, subscribeEvents)
	if err != nil {
		return nil, fmt.Errorf("insert bot: %w", err)
	}

	_, err = d.db.Exec(ctx,
		`INSERT INTO bot_credential (bot_id, client_id, client_secret) VALUES ($1, $2, $3)`,
		botID, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("insert credential: %w", err)
	}

	_, err = d.db.Exec(ctx,
		`INSERT INTO bot_rate_limit (bot_id) VALUES ($1)`, botID)
	if err != nil {
		return nil, fmt.Errorf("insert rate_limit: %w", err)
	}

	return &bot.CreateBotResp{
		BotId:        botID,
		ClientId:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func (d *BotDao) GetBotByID(ctx context.Context, botID, uid int64) (*bot.BotInfo, error) {
	var name, avatar, description, webhookURL, status string
	var ownerUID, createdAt int64
	var subscribeEvents []string

	row := d.db.QueryRow(ctx,
		`SELECT bot_id, name, avatar, description, owner_uid, webhook_url,
		        subscribe_events, status, EXTRACT(EPOCH FROM created_at)::bigint
		 FROM bot WHERE bot_id = $1 AND owner_uid = $2`, botID, uid)
	if err := row.Scan(&botID, &name, &avatar, &description, &ownerUID, &webhookURL,
		&subscribeEvents, &status, &createdAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("bot not found")
		}
		return nil, err
	}

	var events []string = subscribeEvents

	var clientID string
	d.db.QueryRow(ctx,
		`SELECT client_id FROM bot_credential WHERE bot_id = $1`, botID).Scan(&clientID)

	return &bot.BotInfo{
		BotId:           botID,
		Name:            name,
		Avatar:          avatar,
		Description:     description,
		OwnerUid:        ownerUID,
		WebhookUrl:      webhookURL,
		SubscribeEvents: events,
		Status:          status,
		ClientId:        clientID,
		CreatedAt:       createdAt,
	}, nil
}

func (d *BotDao) GetBotByClientID(ctx context.Context, clientID string) (*bot.BotInfo, error) {
	var botID, ownerUID, createdAt int64
	var name, avatar, description, status string
	var subscribeEvents []string

	row := d.db.QueryRow(ctx,
		`SELECT b.bot_id, b.name, b.avatar, b.description, b.owner_uid,
		        b.subscribe_events, b.status, EXTRACT(EPOCH FROM b.created_at)::bigint
		 FROM bot b
		 JOIN bot_credential bc ON b.bot_id = bc.bot_id
		 WHERE bc.client_id = $1 AND b.status = 'ACTIVE'`, clientID)
	if err := row.Scan(&botID, &name, &avatar, &description, &ownerUID,
		&subscribeEvents, &status, &createdAt); err != nil {
		return nil, fmt.Errorf("bot not found")
	}

	return &bot.BotInfo{
		BotId:           botID,
		Name:            name,
		Avatar:          avatar,
		Description:     description,
		OwnerUid:        ownerUID,
		SubscribeEvents: subscribeEvents,
		Status:          status,
		ClientId:        clientID,
		CreatedAt:       createdAt,
	}, nil
}

func (d *BotDao) GetBotCredentials(ctx context.Context, clientID string) (botID int64, clientSecret string, err error) {
	row := d.db.QueryRow(ctx,
		`SELECT b.bot_id, bc.client_secret FROM bot b
		 JOIN bot_credential bc ON b.bot_id = bc.bot_id
		 WHERE bc.client_id = $1 AND b.status = 'ACTIVE'`, clientID)
	err = row.Scan(&botID, &clientSecret)
	return
}

func (d *BotDao) CheckBotOwnership(ctx context.Context, botID, uid int64) error {
	var ownerUID int64
	row := d.db.QueryRow(ctx,
		`SELECT owner_uid FROM bot WHERE bot_id = $1`, botID)
	if err := row.Scan(&ownerUID); err != nil || ownerUID != uid {
		return fmt.Errorf("bot not found or not authorized")
	}
	return nil
}

func (d *BotDao) ListBotsByOwner(ctx context.Context, uid int64) ([]*bot.BotInfo, error) {
	rows, err := d.db.Query(ctx,
		`SELECT b.bot_id, b.name,
		        COALESCE(b.avatar, ''),
		        COALESCE(b.description, ''),
		        b.owner_uid,
		        COALESCE(b.webhook_url, ''),
		        b.subscribe_events, b.status,
		        EXTRACT(EPOCH FROM b.created_at)::bigint,
		        COALESCE(bc.client_id, '')
		 FROM bot b
		 LEFT JOIN bot_credential bc ON b.bot_id = bc.bot_id
		 WHERE b.owner_uid = $1
		 ORDER BY b.bot_id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*bot.BotInfo
	for rows.Next() {
		var botID, ownerUID, createdAt int64
		var name, avatar, description, webhookURL, status, clientID string
		var subscribeEvents []string
		if err := rows.Scan(&botID, &name, &avatar, &description, &ownerUID,
			&webhookURL, &subscribeEvents, &status, &createdAt, &clientID); err != nil {
			continue
		}

		list = append(list, &bot.BotInfo{
			BotId:           botID,
			Name:            name,
			Avatar:          avatar,
			Description:     description,
			OwnerUid:        ownerUID,
			WebhookUrl:      webhookURL,
			SubscribeEvents: subscribeEvents,
			Status:          status,
			ClientId:        clientID,
			CreatedAt:       createdAt,
		})
	}

	return list, nil
}

func (d *BotDao) ListCommunityBots(ctx context.Context, keyword string) ([]*bot.BotInfo, error) {
	var rows pgx.Rows
	var err error

	kw := strings.TrimSpace(keyword)

	if len(kw) == 0 {
		// No keyword: return only official bots
		rows, err = d.db.Query(ctx,
			`SELECT b.bot_id, b.name,
			        COALESCE(b.avatar, ''),
			        COALESCE(b.description, ''),
			        b.owner_uid,
			        COALESCE(b.webhook_url, ''),
			        b.subscribe_events, b.status,
			        EXTRACT(EPOCH FROM b.created_at)::bigint,
			        COALESCE(bc.client_id, '')
			 FROM bot b
			 LEFT JOIN bot_credential bc ON b.bot_id = bc.bot_id
			 WHERE b.owner_uid = 0 AND b.status = 'ACTIVE'
			 ORDER BY b.bot_id DESC`)
	} else {
		// With keyword: search by name (include official + community bots)
		rows, err = d.db.Query(ctx,
			`SELECT b.bot_id, b.name,
			        COALESCE(b.avatar, ''),
			        COALESCE(b.description, ''),
			        b.owner_uid,
			        COALESCE(b.webhook_url, ''),
			        b.subscribe_events, b.status,
			        EXTRACT(EPOCH FROM b.created_at)::bigint,
			        COALESCE(bc.client_id, '')
			 FROM bot b
			 LEFT JOIN bot_credential bc ON b.bot_id = bc.bot_id
			 WHERE (b.owner_uid = 0 OR b.owner_uid > 0) AND b.status = 'ACTIVE'
			   AND b.name ILIKE $1
			 ORDER BY b.bot_id DESC`, "%"+kw+"%")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*bot.BotInfo
	for rows.Next() {
		var botID, ownerUID, createdAt int64
		var name, avatar, description, webhookURL, status, clientID string
		var subscribeEvents []string
		if err := rows.Scan(&botID, &name, &avatar, &description, &ownerUID,
			&webhookURL, &subscribeEvents, &status, &createdAt, &clientID); err != nil {
			continue
		}

		list = append(list, &bot.BotInfo{
			BotId:           botID,
			Name:            name,
			Avatar:          avatar,
			Description:     description,
			OwnerUid:        ownerUID,
			WebhookUrl:      webhookURL,
			SubscribeEvents: subscribeEvents,
			Status:          status,
			ClientId:        clientID,
			CreatedAt:       createdAt,
		})
	}

	return list, nil
}

func (d *BotDao) UpdateBot(ctx context.Context, botID, uid int64, updates map[string]interface{}) error {
	setClauses := []string{}
	args := []interface{}{botID}
	argIdx := 2

	if v, ok := updates["name"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["avatar"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("avatar = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["description"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["webhook_url"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("webhook_url = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["subscribe_events"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("subscribe_events = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	sqlStr := "UPDATE bot SET "
	for i, clause := range setClauses {
		if i > 0 {
			sqlStr += ", "
		}
		sqlStr += clause
	}
	sqlStr += fmt.Sprintf(" WHERE bot_id = $1 AND owner_uid = $%d", argIdx)
	args = append(args, uid)

	_, err := d.db.Exec(ctx, sqlStr, args...)
	return err
}

func (d *BotDao) DeleteBot(ctx context.Context, botID, uid int64) (int64, error) {
	result, err := d.db.Exec(ctx,
		`UPDATE bot SET status = 'DELETED' WHERE bot_id = $1 AND owner_uid = $2`,
		botID, uid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (d *BotDao) RegenerateCredential(ctx context.Context, botID int64) (clientID, clientSecret string, err error) {
	clientID = "bot_" + randomHex(16)
	clientSecret = randomHex(32)

	_, err = d.db.Exec(ctx,
		`UPDATE bot_credential SET client_id = $1, client_secret = $2 WHERE bot_id = $3`,
		clientID, clientSecret, botID)
	return
}

func (d *BotDao) UpdateWebhookSecret(ctx context.Context, botID int64, secret string) error {
	_, err := d.db.Exec(ctx,
		`UPDATE bot SET webhook_secret = $1 WHERE bot_id = $2`,
		secret, botID)
	return err
}

func (d *BotDao) GetRateLimit(ctx context.Context, botID int64) (int64, int64, int64, error) {
	var messagesPerMinute, messagesPerDay, apiCallsPerMinute int
	row := d.db.QueryRow(ctx,
		`SELECT messages_per_minute, messages_per_day, api_calls_per_minute
		 FROM bot_rate_limit WHERE bot_id = $1`, botID)
	err := row.Scan(&messagesPerMinute, &messagesPerDay, &apiCallsPerMinute)
	return int64(messagesPerMinute), int64(messagesPerDay), int64(apiCallsPerMinute), err
}

func (d *BotDao) UpdateRateLimit(ctx context.Context, botID int64, updates map[string]int) error {
	setClauses := []string{}
	args := []interface{}{botID}
	argIdx := 2

	if v, ok := updates["messages_per_minute"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("messages_per_minute = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["messages_per_day"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("messages_per_day = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}
	if v, ok := updates["api_calls_per_minute"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("api_calls_per_minute = $%d", argIdx))
		args = append(args, v)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	sqlStmt := "UPDATE bot_rate_limit SET "
	for i, clause := range setClauses {
		if i > 0 {
			sqlStmt += ", "
		}
		sqlStmt += clause
	}
	sqlStmt += " WHERE bot_id = $1"

	_, err := d.db.Exec(ctx, sqlStmt, args...)
	return err
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func HashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}
