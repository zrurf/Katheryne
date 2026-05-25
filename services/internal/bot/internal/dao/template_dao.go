package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bot/bot"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateDao struct {
	db *pgxpool.Pool
}

func NewTemplateDao(db *pgxpool.Pool) *TemplateDao {
	return &TemplateDao{db: db}
}

func (d *TemplateDao) CreateTemplate(ctx context.Context, req *bot.CreateBotTemplateReq) (int64, error) {
	templateID := time.Now().UnixNano()

	_, err := d.db.Exec(ctx,
		`INSERT INTO bot_template (template_id, name, avatar, description, owner_uid,
		 category, system_prompt, welcome_message, conversation_style, tool_definitions,
		 kb_structure, config_schema, supported_models, tags, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb,
		         $11::jsonb, $12::jsonb, $13::jsonb, $14, 'DRAFT')`,
		templateID, req.Name, req.Avatar, req.Description, req.OwnerUid,
		req.Category, req.SystemPrompt, req.WelcomeMessage,
		req.ConversationStyle, req.ToolDefinitions, req.KbStructure,
		req.ConfigSchema, req.SupportedModels, req.Tags)
	if err != nil {
		return 0, fmt.Errorf("insert template: %w", err)
	}
	return templateID, nil
}

func (d *TemplateDao) GetTemplateByID(ctx context.Context, templateID int64) (*bot.BotTemplateInfo, error) {
	var t bot.BotTemplateInfo
	var createdAt, updatedAt int64

	err := d.db.QueryRow(ctx,
		`SELECT template_id, name, COALESCE(avatar,''), COALESCE(description,''),
		        owner_uid, category, version, system_prompt,
		        COALESCE(welcome_message,''), conversation_style::text, tool_definitions::text,
		        kb_structure::text, config_schema::text, supported_models::text,
		        is_official, tags, status,
		        EXTRACT(EPOCH FROM created_at)::bigint,
		        EXTRACT(EPOCH FROM updated_at)::bigint
		 FROM bot_template WHERE template_id = $1`, templateID).
		Scan(&t.TemplateId, &t.Name, &t.Avatar, &t.Description,
			&t.OwnerUid, &t.Category, &t.Version, &t.SystemPrompt,
			&t.WelcomeMessage, &t.ConversationStyle, &t.ToolDefinitions,
			&t.KbStructure, &t.ConfigSchema, &t.SupportedModels,
			&t.IsOfficial, &t.Tags, &t.Status, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, err
	}
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	return &t, nil
}

func (d *TemplateDao) UpdateTemplate(ctx context.Context, uid, templateID int64, req *bot.UpdateBotTemplateReq) error {
	result, err := d.db.Exec(ctx,
		`UPDATE bot_template SET
		 name = COALESCE(NULLIF($3,''), name),
		 avatar = COALESCE(NULLIF($4,''), avatar),
		 description = COALESCE(NULLIF($5,''), description),
		 category = COALESCE(NULLIF($6,''), category),
		 system_prompt = COALESCE(NULLIF($7,''), system_prompt),
		 welcome_message = COALESCE(NULLIF($8,''), welcome_message),
		 conversation_style = CASE WHEN $9 = '' THEN conversation_style ELSE $9::jsonb END,
		 tool_definitions = CASE WHEN $10 = '' THEN tool_definitions ELSE $10::jsonb END,
		 kb_structure = CASE WHEN $11 = '' THEN kb_structure ELSE $11::jsonb END,
		 config_schema = CASE WHEN $12 = '' THEN config_schema ELSE $12::jsonb END,
		 supported_models = CASE WHEN $13 = '' THEN supported_models ELSE $13::jsonb END,
		 tags = CASE WHEN $14 IS NULL OR cardinality($14) = 0 THEN tags ELSE $14 END,
		 updated_at = NOW()
		 WHERE template_id = $2 AND owner_uid = $1`,
		uid, templateID, req.Name, req.Avatar, req.Description, req.Category,
		req.SystemPrompt, req.WelcomeMessage, req.ConversationStyle, req.ToolDefinitions,
		req.KbStructure, req.ConfigSchema, req.SupportedModels, req.Tags)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found or not authorized")
	}
	return nil
}

func (d *TemplateDao) DeleteTemplate(ctx context.Context, uid, templateID int64) error {
	result, err := d.db.Exec(ctx,
		`UPDATE bot_template SET status = 'DELETED', updated_at = NOW()
		 WHERE template_id = $1 AND owner_uid = $2 AND status != 'DELETED'`,
		templateID, uid)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found or not authorized")
	}
	return nil
}

func (d *TemplateDao) ListTemplatesByOwner(ctx context.Context, ownerUID int64) ([]*bot.BotTemplateInfo, error) {
	rows, err := d.db.Query(ctx,
		`SELECT template_id, name, COALESCE(avatar,''), COALESCE(description,''),
		        owner_uid, category, version, system_prompt,
		        COALESCE(welcome_message,''), conversation_style::text, tool_definitions::text,
		        kb_structure::text, config_schema::text, supported_models::text,
		        is_official, tags, status,
		        EXTRACT(EPOCH FROM created_at)::bigint,
		        EXTRACT(EPOCH FROM updated_at)::bigint
		 FROM bot_template WHERE owner_uid = $1 AND status != 'DELETED'
		 ORDER BY created_at DESC`, ownerUID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var list []*bot.BotTemplateInfo
	for rows.Next() {
		var t bot.BotTemplateInfo
		var createdAt, updatedAt int64
		if err := rows.Scan(&t.TemplateId, &t.Name, &t.Avatar, &t.Description,
			&t.OwnerUid, &t.Category, &t.Version, &t.SystemPrompt,
			&t.WelcomeMessage, &t.ConversationStyle, &t.ToolDefinitions,
			&t.KbStructure, &t.ConfigSchema, &t.SupportedModels,
			&t.IsOfficial, &t.Tags, &t.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		list = append(list, &t)
	}
	return list, nil
}

func (d *TemplateDao) PublishTemplate(ctx context.Context, uid, templateID int64) error {
	result, err := d.db.Exec(ctx,
		`UPDATE bot_template SET status = 'PUBLISHED', updated_at = NOW()
		 WHERE template_id = $1 AND owner_uid = $2 AND status = 'DRAFT'`,
		templateID, uid)
	if err != nil {
		return fmt.Errorf("publish template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found, not authorized, or already published")
	}
	return nil
}

// ListCommunityTemplates lists published templates for the marketplace
func (d *TemplateDao) ListCommunityTemplates(ctx context.Context, keyword, category string) ([]*bot.BotTemplateInfo, error) {
	query := `SELECT template_id, name, COALESCE(avatar,''), COALESCE(description,''),
		          owner_uid, category, version, system_prompt,
		          COALESCE(welcome_message,''), conversation_style::text, tool_definitions::text,
		          kb_structure::text, config_schema::text, supported_models::text,
		          is_official, tags, status,
		          EXTRACT(EPOCH FROM created_at)::bigint,
		          EXTRACT(EPOCH FROM updated_at)::bigint
		   FROM bot_template WHERE status = 'PUBLISHED'`

	args := []interface{}{}
	argIdx := 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}
	if keyword != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR $%d = ANY(tags))", argIdx, argIdx+1, argIdx+2)
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, keyword)
		argIdx += 3
	}

	query += " ORDER BY is_official DESC, display_order ASC, created_at DESC"

	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list community templates: %w", err)
	}
	defer rows.Close()

	var list []*bot.BotTemplateInfo
	for rows.Next() {
		var t bot.BotTemplateInfo
		var createdAt, updatedAt int64
		if err := rows.Scan(&t.TemplateId, &t.Name, &t.Avatar, &t.Description,
			&t.OwnerUid, &t.Category, &t.Version, &t.SystemPrompt,
			&t.WelcomeMessage, &t.ConversationStyle, &t.ToolDefinitions,
			&t.KbStructure, &t.ConfigSchema, &t.SupportedModels,
			&t.IsOfficial, &t.Tags, &t.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		list = append(list, &t)
	}
	return list, nil
}

// unmarshalJSON is a helper for logic files
func unmarshalJSON(s string) interface{} {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}