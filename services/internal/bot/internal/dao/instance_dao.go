package dao

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"bot/bot"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const encryptKey = "katheryne-bot-instance-key-32b!" // 32 bytes for AES-256

type InstanceDao struct {
	db          *pgxpool.Pool
	templateDao *TemplateDao
}

func NewInstanceDao(db *pgxpool.Pool, templateDao *TemplateDao) *InstanceDao {
	return &InstanceDao{db: db, templateDao: templateDao}
}

func (d *InstanceDao) CreateInstance(ctx context.Context, req *bot.CreateBotInstanceReq) (*bot.CreateBotInstanceResp, error) {
	instanceID := time.Now().UnixNano()
	botID := time.Now().UnixNano() + 1
	clientID := "bot_" + randomHex(16)
	clientSecret := randomHex(32)
	webhookSecret := randomHex(16)
	botToken := "bt_" + randomHex(32)

	// Encrypt API key
	encryptedKey, err := encryptAPIKey(req.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt api key: %w", err)
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create bot identity
	_, err = tx.Exec(ctx,
		`INSERT INTO bot (bot_id, name, avatar, description, owner_uid, webhook_secret, subscribe_events, status)
		 VALUES ($1, $2, $3, '', $4, $5, '{}', 'ACTIVE')`,
		botID, req.Name, req.Avatar, req.Uid, webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("insert bot: %w", err)
	}

	// Create bot credential
	_, err = tx.Exec(ctx,
		`INSERT INTO bot_credential (bot_id, client_id, client_secret) VALUES ($1, $2, $3)`,
		botID, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("insert credential: %w", err)
	}

	// Create bot rate limit
	_, err = tx.Exec(ctx, `INSERT INTO bot_rate_limit (bot_id) VALUES ($1)`, botID)
	if err != nil {
		return nil, fmt.Errorf("insert rate limit: %w", err)
	}

	// Normalize KB config
	kbConfig := req.KbConfig
	if kbConfig == "" {
		kbConfig = "{}"
	}
	instanceConfig := req.InstanceConfig
	if instanceConfig == "" {
		instanceConfig = "{}"
	}

	// Determine hosted_by
	hostedBy := req.HostedBy
	if !req.IsSelfHosted && hostedBy == 0 {
		hostedBy = int64(0) // official
	}

	// Create instance
	_, err = tx.Exec(ctx,
		`INSERT INTO bot_instance (instance_id, bot_id, template_id, owner_uid,
		 name, avatar, is_self_hosted, hosted_by, model_provider, model_name,
		 api_key_encrypted, api_base_url, kb_config, instance_config,
		 bot_token, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		         $11, $12, $13::jsonb, $14::jsonb, $15, 'ACTIVE')`,
		instanceID, botID, req.TemplateId, req.Uid,
		req.Name, req.Avatar, req.IsSelfHosted, hostedBy,
		req.ModelProvider, req.ModelName,
		encryptedKey, req.ApiBaseUrl,
		kbConfig, instanceConfig, botToken)
	if err != nil {
		return nil, fmt.Errorf("insert instance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &bot.CreateBotInstanceResp{
		InstanceId:   instanceID,
		BotId:        botID,
		ClientId:     clientID,
		ClientSecret: clientSecret,
		BotToken:     botToken,
	}, nil
}

func (d *InstanceDao) GetInstanceByID(ctx context.Context, instanceID int64) (*bot.BotInstanceInfo, error) {
	return d.queryInstance(ctx, "SELECT instance_id FROM bot_instance WHERE instance_id = $1 AND status != 'DELETED'", instanceID)
}

func (d *InstanceDao) GetInstanceByBotID(ctx context.Context, botID int64) (*bot.BotInstanceInfo, error) {
	return d.queryInstance(ctx, "SELECT instance_id FROM bot_instance WHERE bot_id = $1 AND status != 'DELETED'", botID)
}

func (d *InstanceDao) queryInstance(ctx context.Context, idQuery string, id int64) (*bot.BotInstanceInfo, error) {
	var inst bot.BotInstanceInfo
	var kbConfig, instConfig string
	var createTs int64

	err := d.db.QueryRow(ctx,
		`SELECT instance_id, bot_id, template_id, owner_uid,
		        name, COALESCE(avatar,''), is_self_hosted, COALESCE(hosted_by,0),
		        COALESCE(model_provider,''), COALESCE(model_name,''),
		        kb_config::text, instance_config::text, status,
		        EXTRACT(EPOCH FROM created_at)::bigint
		 FROM bot_instance WHERE `+idQuery, id).
		Scan(&inst.InstanceId, &inst.BotId, &inst.TemplateId, &inst.OwnerUid,
			&inst.Name, &inst.Avatar, &inst.IsSelfHosted, &inst.HostedBy,
			&inst.ModelProvider, &inst.ModelName,
			&kbConfig, &instConfig, &inst.Status, &createTs)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("instance not found")
		}
		return nil, err
	}
	inst.KbConfig = kbConfig
	inst.InstanceConfig = instConfig
	inst.CreatedAt = createTs

	// Load template
	tmpl, err := d.templateDao.GetTemplateByID(ctx, inst.TemplateId)
	if err == nil {
		inst.Template = tmpl
	}

	return &inst, nil
}

func (d *InstanceDao) UpdateInstance(ctx context.Context, uid, instanceID int64, req *bot.UpdateBotInstanceReq) error {
	// Only allow the owner to update
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify ownership
	var ownerUID int64
	err = tx.QueryRow(ctx,
		`SELECT owner_uid FROM bot_instance WHERE instance_id = $1 AND status != 'DELETED'`,
		instanceID).Scan(&ownerUID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}
	if ownerUID != uid {
		return fmt.Errorf("not authorized")
	}

	// Update fields
	if req.ApiKey != "" {
		encryptedKey, err := encryptAPIKey(req.ApiKey)
		if err != nil {
			return fmt.Errorf("encrypt api key: %w", err)
		}
		_, err = tx.Exec(ctx,
			`UPDATE bot_instance SET api_key_encrypted = $1, updated_at = NOW()
			 WHERE instance_id = $2`, encryptedKey, instanceID)
		if err != nil {
			return fmt.Errorf("update api key: %w", err)
		}
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.ModelProvider != "" {
		updates["model_provider"] = req.ModelProvider
		updates["bot_model"] = req.ModelProvider // also update bot table
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
	}
	if req.ApiBaseUrl != "" {
		updates["api_base_url"] = req.ApiBaseUrl
	}
	if req.KbConfig != "" {
		updates["kb_config"] = req.KbConfig
	}
	if req.InstanceConfig != "" {
		updates["instance_config"] = req.InstanceConfig
	}

	if len(updates) > 0 {
		query := "UPDATE bot_instance SET updated_at = NOW()"
		args := []interface{}{}
		argIdx := 1
		for col, val := range updates {
			query += fmt.Sprintf(", %s = $%d", col, argIdx)
			args = append(args, val)
			argIdx++
		}
		query += fmt.Sprintf(" WHERE instance_id = $%d", argIdx)
		args = append(args, instanceID)
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("update instance: %w", err)
		}
	}

	// Update bot name/avatar if changed
	if req.Name != "" || req.Avatar != "" {
		var botID int64
		tx.QueryRow(ctx, `SELECT bot_id FROM bot_instance WHERE instance_id = $1`, instanceID).Scan(&botID)
		if req.Name != "" && req.Avatar != "" {
			tx.Exec(ctx, `UPDATE bot SET name = $1, avatar = $2, updated_at = NOW() WHERE bot_id = $3`,
				req.Name, req.Avatar, botID)
		} else if req.Name != "" {
			tx.Exec(ctx, `UPDATE bot SET name = $1, updated_at = NOW() WHERE bot_id = $2`,
				req.Name, botID)
		} else if req.Avatar != "" {
			tx.Exec(ctx, `UPDATE bot SET avatar = $1, updated_at = NOW() WHERE bot_id = $2`,
				req.Avatar, botID)
		}
	}

	return tx.Commit(ctx)
}

func (d *InstanceDao) DeleteInstance(ctx context.Context, uid, instanceID int64) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if instance exists at all (including already deleted)
	var ownerUID, botID int64
	var status string
	err = tx.QueryRow(ctx,
		`SELECT owner_uid, bot_id, status FROM bot_instance WHERE instance_id = $1`,
		instanceID).Scan(&ownerUID, &botID, &status)
	if err != nil {
		return fmt.Errorf("instance not found")
	}

	// Check if already deleted
	if status == "DELETED" {
		return fmt.Errorf("instance already deleted")
	}

	// Verify ownership
	if ownerUID != uid {
		return fmt.Errorf("not authorized to delete this instance")
	}

	// Soft delete instance
	_, err = tx.Exec(ctx,
		`UPDATE bot_instance SET status = 'DELETED', updated_at = NOW() WHERE instance_id = $1`, instanceID)
	if err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}

	// Soft delete bot identity
	_, err = tx.Exec(ctx,
		`UPDATE bot SET status = 'DELETED', updated_at = NOW() WHERE bot_id = $1`, botID)
	if err != nil {
		return fmt.Errorf("delete bot: %w", err)
	}

	return tx.Commit(ctx)
}

func (d *InstanceDao) ListInstancesByOwner(ctx context.Context, ownerUID int64) ([]*bot.BotInstanceInfo, error) {
	rows, err := d.db.Query(ctx,
		`SELECT instance_id, bot_id, template_id, owner_uid,
		        name, COALESCE(avatar,''), is_self_hosted, COALESCE(hosted_by,0),
		        COALESCE(model_provider,''), COALESCE(model_name,''),
		        kb_config::text, instance_config::text, status,
		        EXTRACT(EPOCH FROM created_at)::bigint
		 FROM bot_instance WHERE owner_uid = $1 AND status != 'DELETED'
		 ORDER BY created_at DESC`, ownerUID)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	var list []*bot.BotInstanceInfo
	for rows.Next() {
		var inst bot.BotInstanceInfo
		var kbConfig, instConfig string
		var createTs int64
		if err := rows.Scan(&inst.InstanceId, &inst.BotId, &inst.TemplateId, &inst.OwnerUid,
			&inst.Name, &inst.Avatar, &inst.IsSelfHosted, &inst.HostedBy,
			&inst.ModelProvider, &inst.ModelName,
			&kbConfig, &instConfig, &inst.Status, &createTs); err != nil {
			return nil, err
		}
		inst.KbConfig = kbConfig
		inst.InstanceConfig = instConfig
		inst.CreatedAt = createTs
		// Load template info
		tmpl, err := d.templateDao.GetTemplateByID(ctx, inst.TemplateId)
		if err == nil {
			inst.Template = tmpl
		}
		list = append(list, &inst)
	}
	return list, nil
}

// ListHostedInstances lists non-self-hosted instances for the community marketplace
func (d *InstanceDao) ListHostedInstances(ctx context.Context, keyword string) ([]*bot.BotInstanceInfo, error) {
	query := `SELECT bi.instance_id, bi.bot_id, bi.template_id, bi.owner_uid,
		          bi.name, COALESCE(bi.avatar,''), bi.is_self_hosted, COALESCE(bi.hosted_by,0),
		          COALESCE(bi.model_provider,''), COALESCE(bi.model_name,''),
		          bi.kb_config::text, bi.instance_config::text, bi.status,
		          EXTRACT(EPOCH FROM bi.created_at)::bigint
		   FROM bot_instance bi
		   WHERE bi.status = 'ACTIVE'
		     AND (bi.is_self_hosted = FALSE OR bi.hosted_by = 0)`

	args := []interface{}{}
	argIdx := 1

	if keyword != "" {
		query += fmt.Sprintf(" AND (bi.name ILIKE $%d)", argIdx)
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	query += " ORDER BY bi.hosted_by = 0 DESC, bi.created_at DESC"

	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list hosted instances: %w", err)
	}
	defer rows.Close()

	var list []*bot.BotInstanceInfo
	for rows.Next() {
		var inst bot.BotInstanceInfo
		var kbConfig, instConfig string
		var createTs int64
		if err := rows.Scan(&inst.InstanceId, &inst.BotId, &inst.TemplateId, &inst.OwnerUid,
			&inst.Name, &inst.Avatar, &inst.IsSelfHosted, &inst.HostedBy,
			&inst.ModelProvider, &inst.ModelName,
			&kbConfig, &instConfig, &inst.Status, &createTs); err != nil {
			return nil, err
		}
		inst.KbConfig = kbConfig
		inst.InstanceConfig = instConfig
		inst.CreatedAt = createTs
		// Load template info
		tmpl, err := d.templateDao.GetTemplateByID(ctx, inst.TemplateId)
		if err == nil {
			inst.Template = tmpl
		}
		list = append(list, &inst)
	}
	return list, nil
}

// VerifyBotToken checks if a bot_token is valid and returns bot info
func (d *InstanceDao) VerifyBotToken(ctx context.Context, token string) (botID, instanceID int64, isOfficial bool, err error) {
	row := d.db.QueryRow(ctx,
		`SELECT bot_id, instance_id, hosted_by = 0
		 FROM bot_instance WHERE bot_token = $1 AND status = 'ACTIVE'`, token)
	err = row.Scan(&botID, &instanceID, &isOfficial)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid bot token")
	}
	return
}

// GetRuntimeConfig loads the bot's runtime configuration for LLM calls
func (d *InstanceDao) GetRuntimeConfig(ctx context.Context, botID, convID int64) (*bot.GetBotRuntimeConfigResp, error) {
	var resp bot.GetBotRuntimeConfigResp
	var encryptedKey, apiBase, kbConfig string

	resp.BotId = botID

	row := d.db.QueryRow(ctx,
		`SELECT bi.instance_id, COALESCE(bi.model_provider,''), COALESCE(bi.model_name,''),
		        COALESCE(bi.api_key_encrypted,''), COALESCE(bi.api_base_url,''),
		        bi.kb_config::text, bi.hosted_by = 0
		 FROM bot_instance bi
		 WHERE bi.bot_id = $1 AND bi.status = 'ACTIVE'`, botID)
	err := row.Scan(&resp.InstanceId, &resp.ModelProvider, &resp.ModelName,
		&encryptedKey, &apiBase, &kbConfig, &resp.IsOfficial)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	// Decrypt API key
	if encryptedKey != "" {
		decrypted, err := decryptAPIKey(encryptedKey)
		if err == nil {
			resp.ApiKey = decrypted
		}
	}
	resp.ApiBaseUrl = apiBase
	resp.KbConfig = kbConfig

	// Load credential (client_id, client_secret)
	var clientID, clientSecret string
	row3 := d.db.QueryRow(ctx,
		`SELECT client_id, client_secret FROM bot_credential WHERE bot_id = $1`, botID)
	if err := row3.Scan(&clientID, &clientSecret); err == nil {
		resp.ClientId = clientID
		resp.ClientSecret = clientSecret
	}

	// Load bot name
	var botName string
	row4 := d.db.QueryRow(ctx,
		`SELECT name FROM bot WHERE bot_id = $1`, botID)
	if err := row4.Scan(&botName); err == nil {
		resp.Name = botName
	}

	// Load template info
	var templateID int64
	row2 := d.db.QueryRow(ctx,
		`SELECT template_id FROM bot_instance WHERE bot_id = $1`, botID)
	if err := row2.Scan(&templateID); err == nil {
		tmpl, err := d.templateDao.GetTemplateByID(ctx, templateID)
		if err == nil {
			resp.SystemPrompt = tmpl.SystemPrompt
			resp.ConversationStyle = tmpl.ConversationStyle
			resp.ToolDefinitions = tmpl.ToolDefinitions
		}
	}

	// Load authorized KB IDs for this conversation
	if convID > 0 {
		rows, err := d.db.Query(ctx,
			`SELECT ka.kb_id FROM kb_auth ka
			 INNER JOIN bot_instance bi ON bi.bot_id = ka.bot_id
			 WHERE bi.bot_id = $1 AND ka.conv_id = $2`, botID, convID)
		if err == nil {
			defer rows.Close()
			var kbIDs []string
			for rows.Next() {
				var kbID string
				if rows.Scan(&kbID) == nil {
					kbIDs = append(kbIDs, kbID)
				}
			}
			resp.KbIds = strings.Join(kbIDs, ",")
		}
	} else {
		// Bot-level: fetch all KB IDs this bot has access to (across all conversations)
		rows, err := d.db.Query(ctx,
			`SELECT DISTINCT ka.kb_id FROM kb_auth ka
			 INNER JOIN bot_instance bi ON bi.bot_id = ka.bot_id
			 WHERE bi.bot_id = $1`, botID)
		if err == nil {
			defer rows.Close()
			var kbIDs []string
			for rows.Next() {
				var kbID string
				if rows.Scan(&kbID) == nil {
					kbIDs = append(kbIDs, kbID)
				}
			}
			resp.KbIds = strings.Join(kbIDs, ",")
		}
	}

	return &resp, nil
}

// encryptAPIKey encrypts sensitive data using AES-GCM
func encryptAPIKey(plaintext string) (string, error) {
	key := sha256.Sum256([]byte(encryptKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decryptAPIKey decrypts AES-GCM encrypted data
func decryptAPIKey(cipherHex string) (string, error) {
	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(encryptKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
