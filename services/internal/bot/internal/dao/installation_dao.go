package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bot/bot"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InstallationDao struct {
	db       *pgxpool.Pool
	socialDB *pgxpool.Pool
}

func NewInstallationDao(db, socialDB *pgxpool.Pool) *InstallationDao {
	return &InstallationDao{db: db, socialDB: socialDB}
}

func (d *InstallationDao) GetConversation(ctx context.Context, convID int64, uid int64) (convType string, groupID int64, err error) {
	if d.socialDB == nil {
		return "", 0, fmt.Errorf("social database not configured")
	}
	row := d.socialDB.QueryRow(ctx,
		`SELECT type, COALESCE(group_id, 0) FROM conversations WHERE conv_id = $1`, convID)
	err = row.Scan(&convType, &groupID)
	if err != nil {
		err = fmt.Errorf("conversation not found")
	}
	return
}

func (d *InstallationDao) CheckGroupMemberRole(ctx context.Context, groupID, uid int64, allowedRoles ...string) error {
	db := d.socialDB
	if db == nil {
		db = d.db
	}
	var role string
	row := db.QueryRow(ctx,
		`SELECT role FROM group_members WHERE group_id = $1 AND uid = $2`, groupID, uid)
	if err := row.Scan(&role); err != nil {
		return fmt.Errorf("not a member of this group")
	}
	for _, r := range allowedRoles {
		if role == r {
			return nil
		}
	}
	return fmt.Errorf("insufficient group role: %s", role)
}

func (d *InstallationDao) IsGroupMember(ctx context.Context, groupID, uid int64) bool {
	var count int
	row := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM im_group_members WHERE group_id = $1 AND uid = $2`, groupID, uid)
	row.Scan(&count)
	return count > 0
}

func (d *InstallationDao) IsInstalled(ctx context.Context, botID, convID int64) bool {
	var count int
	row := d.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bot_installation WHERE bot_id = $1 AND conv_id = $2 AND status = 'ACTIVE'`,
		botID, convID)
	row.Scan(&count)
	return count > 0
}

func (d *InstallationDao) Install(ctx context.Context, botID, convID int64, convType string, permissions []string, installedBy int64) error {
	permissionsJSON, _ := json.Marshal(permissions)

	_, err := d.db.Exec(ctx,
		`INSERT INTO bot_installation (bot_id, conv_id, conv_type, permissions, installed_by, status, installed_at)
		 VALUES ($1, $2, $3, $4, $5, 'ACTIVE', $6)
		 ON CONFLICT (bot_id, conv_id) DO UPDATE SET status = 'ACTIVE', permissions = $4, updated_at = $6`,
		botID, convID, convType, string(permissionsJSON), installedBy, time.Now())
	return err
}

func (d *InstallationDao) Uninstall(ctx context.Context, botID, convID int64) error {
	_, err := d.db.Exec(ctx,
		`UPDATE bot_installation SET status = 'INACTIVE', updated_at = $1
		 WHERE bot_id = $2 AND conv_id = $3`,
		time.Now(), botID, convID)
	return err
}

func (d *InstallationDao) ListConvBots(ctx context.Context, convID int64) ([]*bot.InstalledBotItem, error) {
	rows, err := d.db.Query(ctx,
		`SELECT b.bot_id, b.name, b.avatar, b.description, bi.permissions,
		        EXTRACT(EPOCH FROM bi.installed_at)::bigint
		 FROM bot_installation bi
		 JOIN bot b ON bi.bot_id = b.bot_id
		 WHERE bi.conv_id = $1 AND bi.status = 'ACTIVE' AND b.status = 'ACTIVE'
		 ORDER BY bi.installed_at DESC`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*bot.InstalledBotItem
	for rows.Next() {
		var botID, installedAt int64
		var name, avatar, description, permissions string
		if err := rows.Scan(&botID, &name, &avatar, &description, &permissions, &installedAt); err != nil {
			continue
		}

		var perms []string
		json.Unmarshal([]byte(permissions), &perms)

		list = append(list, &bot.InstalledBotItem{
			BotId:       botID,
			Name:        name,
			Avatar:      avatar,
			Description: description,
			Permissions: perms,
			InstalledAt: installedAt,
		})
	}

	return list, nil
}

func (d *InstallationDao) ListBotInstallations(ctx context.Context, botID int64) ([]*bot.BotInstallationItem, error) {
	rows, err := d.db.Query(ctx,
		`SELECT bi.conv_id, bi.conv_type, bi.permissions,
		        EXTRACT(EPOCH FROM bi.installed_at)::bigint
		 FROM bot_installation bi
		 WHERE bi.bot_id = $1 AND bi.status = 'ACTIVE'
		 ORDER BY bi.installed_at DESC`, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*bot.BotInstallationItem
	for rows.Next() {
		var convID, installedAt int64
		var convType, permissionsStr string
		if err := rows.Scan(&convID, &convType, &permissionsStr, &installedAt); err != nil {
			continue
		}

		perms := parsePermissionsStr(permissionsStr)
		list = append(list, &bot.BotInstallationItem{
			ConvId:      convID,
			ConvType:    convType,
			Permissions: perms,
			InstalledAt: installedAt,
		})
	}

	return list, nil
}

func (d *InstallationDao) GetConvInfo(ctx context.Context, convID int64) (convType, name, avatar string, groupID int64, createdAt int64, err error) {
	db := d.socialDB
	if db == nil {
		db = d.db
	}
	var gID *int64
	row := db.QueryRow(ctx,
		`SELECT type, group_id, name, avatar, EXTRACT(EPOCH FROM created_at)::bigint
		 FROM conversations WHERE conv_id = $1`, convID)
	if err = row.Scan(&convType, &gID, &name, &avatar, &createdAt); err != nil {
		return
	}
	if gID != nil {
		groupID = *gID
	}
	return
}

func (d *InstallationDao) GetGroupMembers(ctx context.Context, convID int64) ([]*bot.BotConvMemberItem, error) {
	db := d.socialDB
	if db == nil {
		db = d.db
	}
	rows, err := db.Query(ctx,
		`SELECT gm.uid, gm.role, COALESCE(gm.nick, '')
		 FROM group_members gm
		 WHERE gm.group_id = (SELECT group_id FROM conversations WHERE conv_id = $1)
		 LIMIT 1000`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*bot.BotConvMemberItem
	for rows.Next() {
		var uid int64
		var role, nick string
		if err := rows.Scan(&uid, &role, &nick); err != nil {
			continue
		}
		members = append(members, &bot.BotConvMemberItem{
			Uid:  uid,
			Role: role,
			Nick: nick,
		})
	}

	return members, nil
}

func (d *InstallationDao) GetMessage(ctx context.Context, msgID, convID int64) (*bot.BotGetMsgResp, error) {
	var senderUID, quoteMsgID *int64
	var msgIDOut, cID, createdAt int64
	var senderName, msgType, content, contentType, extra string
	var recalled, edited bool

	row := d.db.QueryRow(ctx,
		`SELECT msg_id, conv_id, sender_uid, sender_name, msg_type, content, content_type,
		        quote_msg_id, recalled, edited, extra, created_at
		 FROM im_message WHERE msg_id = $1 AND conv_id = $2`,
		msgID, convID)
	if err := row.Scan(&msgIDOut, &cID, &senderUID, &senderName, &msgType, &content,
		&contentType, &quoteMsgID, &recalled, &edited, &extra, &createdAt); err != nil {
		return nil, fmt.Errorf("message not found")
	}

	if len(content) > 10000 {
		content = content[:10000]
	}

	resp := &bot.BotGetMsgResp{
		MsgId:       msgIDOut,
		ConvId:      cID,
		SenderName:  senderName,
		MsgType:     msgType,
		Content:     content,
		ContentType: contentType,
		Recalled:    recalled,
		Edited:      edited,
		Extra:       extra,
		CreatedAt:   createdAt,
	}
	if senderUID != nil {
		resp.SenderUid = *senderUID
	}
	if quoteMsgID != nil {
		resp.QuoteMsgId = *quoteMsgID
	}

	return resp, nil
}

func (d *InstallationDao) GetUserInfo(ctx context.Context, uid int64) (name, avatar string, err error) {
	row := d.db.QueryRow(ctx,
		"SELECT name, avatar FROM users WHERE uid = $1", uid)
	err = row.Scan(&name, &avatar)
	return
}

func parsePermissionsStr(permissionsStr string) []string {
	if permissionsStr == "" || permissionsStr == "[]" {
		return nil
	}
	var perms []string
	for i := 0; i < len(permissionsStr); i++ {
		if permissionsStr[i] == '"' {
			end := i + 1
			for end < len(permissionsStr) && permissionsStr[end] != '"' {
				end++
			}
			perms = append(perms, permissionsStr[i+1:end])
			i = end
		}
	}
	return perms
}
