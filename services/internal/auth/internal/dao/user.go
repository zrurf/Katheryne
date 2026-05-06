package dao

import (
	"auth/internal/model"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UpsertUserConfig struct {
	UID             int64
	Language        *string
	MsgNotification *bool
	SoundEnabled    *bool
	AutoTranslate   *bool
	TranslateTarget *string
	ContentFilter   *bool
	Enable2FA       *bool
	TOTPSecret      *string
	TOTPBackupCodes []string
}

type UserDao struct {
	db *pgxpool.Pool
}

func NewUserDao(pool *pgxpool.Pool) *UserDao {
	return &UserDao{db: pool}
}

// ---------- 基础用户操作 ----------

// SaveUserRecord 创建用户并存储 OPAQUE 注册记录。
func (d *UserDao) SaveUserRecord(ctx context.Context, uid int64, phone, name string, record []byte) error {
	log.Debug().
		Int64("uid", uid).
		Str("phone", phone).
		Bytes("record", record).
		Msg("save user record")

	sql := `INSERT INTO users (uid, phone, name, opaque_record, created_at, updated_at, last_login)
	        VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())
	        ON CONFLICT (phone) DO NOTHING`
	_, err := d.db.Exec(ctx, sql, uid, phone, name, record)
	return err
}

// GetUserRecord 根据 phone 查询 uid 和 OPAQUE 记录。
func (d *UserDao) GetUserRecord(ctx context.Context, phone string) (int64, []byte, error) {
	var uid int64
	var record []byte
	sql := `SELECT uid, opaque_record FROM users WHERE phone = $1 LIMIT 1`
	err := d.db.QueryRow(ctx, sql, phone).Scan(&uid, &record)
	log.Debug().Str("phone", phone).Int64("uid", uid).Msg("read user record")
	return uid, record, err
}

// UpdateLastLogin 更新最后登录时间。
func (d *UserDao) UpdateLastLogin(ctx context.Context, uid int64) error {
	sql := `UPDATE users SET last_login = NOW() WHERE uid = $1`
	_, err := d.db.Exec(ctx, sql, uid)
	return err
}

// PhoneExists 检查手机号是否已注册。
func (d *UserDao) PhoneExists(ctx context.Context, phone string) (bool, error) {
	var exists bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 LIMIT 1)`
	err := d.db.QueryRow(ctx, sql, phone).Scan(&exists)
	return exists, err
}

// UidExists 检查 uid 是否已注册
func (d *UserDao) UidExists(ctx context.Context, uid int64) (bool, error) {
	var exists bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE uid = $1 LIMIT 1)`
	err := d.db.QueryRow(ctx, sql, uid).Scan(&exists)
	return exists, err
}

// GetUserByUID 获取用户完整信息。
func (d *UserDao) GetUserByUID(ctx context.Context, uid int64) (*model.User, error) {
	row := d.db.QueryRow(ctx,
		`SELECT uid, name, phone, avatar, status, updated_at, created_at, last_login, opaque_record
		 FROM users WHERE uid = $1`, uid)
	u := &model.User{}
	err := row.Scan(&u.UID, &u.Name, &u.Phone, &u.Avatar, &u.Status,
		&u.UpdatedAt, &u.CreatedAt, &u.LastLogin, &u.OpaqueRecord)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateAvatar 设置用户头像。
func (d *UserDao) UpdateAvatar(ctx context.Context, uid int64, avatar string) error {
	sql := `UPDATE users SET avatar = $1, updated_at = NOW() WHERE uid = $2`
	_, err := d.db.Exec(ctx, sql, avatar, uid)
	return err
}

// UpdateStatus 设置用户状态（需保证状态值符合 CHECK 约束）。
func (d *UserDao) UpdateStatus(ctx context.Context, uid int64, status string) error {
	sql := `UPDATE users SET status = $1, updated_at = NOW() WHERE uid = $2`
	_, err := d.db.Exec(ctx, sql, status, uid)
	return err
}

// ---------- 用户配置表 ----------

// GetUserConfig 获取用户配置，若未初始化则返回默认值。
func (d *UserDao) GetUserConfig(ctx context.Context, uid int64) (*model.UserConfig, error) {
	cfg := &model.UserConfig{}
	err := d.db.QueryRow(ctx,
		`SELECT uid, language, msg_notification, sound_enabled, auto_translate,
		        translate_target, content_filter, enable_2fa, totp_secret,
		        totp_backup_codes, updated_at, created_at
		 FROM user_config WHERE uid = $1`, uid).
		Scan(&cfg.UID, &cfg.Language, &cfg.MsgNotification,
			&cfg.SoundEnabled, &cfg.AutoTranslate, &cfg.TranslateTarget,
			&cfg.ContentFilter, &cfg.Enable2FA, &cfg.TOTPSecret,
			&cfg.TOTPBackupCodes, &cfg.UpdatedAt, &cfg.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			// 没有配置记录，返回 nil 表示未初始化，业务方可视为默认配置
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

// UpsertUserConfig 根据 UpsertUserConfig 部分更新用户配置。
func (d *UserDao) UpsertUserConfig(ctx context.Context, cfg *UpsertUserConfig) error {
	// 1. 确保配置行存在（不存在则插入默认值）
	if err := d.ensureConfigExists(ctx, cfg.UID); err != nil {
		return err
	}

	// 2. 动态构建更新 SQL
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	add := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if cfg.Language != nil {
		add("language", *cfg.Language)
	}
	if cfg.MsgNotification != nil {
		add("msg_notification", *cfg.MsgNotification)
	}
	if cfg.SoundEnabled != nil {
		add("sound_enabled", *cfg.SoundEnabled)
	}
	if cfg.AutoTranslate != nil {
		add("auto_translate", *cfg.AutoTranslate)
	}
	if cfg.TranslateTarget != nil {
		add("translate_target", *cfg.TranslateTarget)
	}
	if cfg.ContentFilter != nil {
		add("content_filter", *cfg.ContentFilter)
	}
	if cfg.Enable2FA != nil {
		add("enable_2fa", *cfg.Enable2FA)
	}
	if cfg.TOTPSecret != nil {
		add("totp_secret", *cfg.TOTPSecret)
	}
	if cfg.TOTPBackupCodes != nil {
		add("totp_backup_codes", cfg.TOTPBackupCodes)
	}

	if len(setClauses) == 0 {
		return nil // 无字段更新
	}

	// 总是刷新 updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE user_config SET %s WHERE uid = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, cfg.UID)

	_, err := d.db.Exec(ctx, query, args...)
	return err
}

// ensureConfigExists 若 target uid 不存在配置行，则插入默认配置。
func (d *UserDao) ensureConfigExists(ctx context.Context, uid int64) error {
	const insertDefault = `INSERT INTO user_config (uid, language, msg_notification, sound_enabled, auto_translate,
		content_filter, enable_2fa, created_at, updated_at)
		VALUES ($1, 'zh-CN', true, true, false, true, false, NOW(), NOW())
		ON CONFLICT (uid) DO NOTHING`
	_, err := d.db.Exec(ctx, insertDefault, uid)
	return err
}

// ---------- 登录日志 ----------

// InsertLoginLog 写入登录日志。
func (d *UserDao) InsertLoginLog(ctx context.Context, logEntry *model.LoginLog) error {
	sql := `INSERT INTO login_log (uid, "user", success, reason, ip, user_agent, created_at)
	        VALUES ($1, $2, $3, $4, $5, $6, NOW())`
	_, err := d.db.Exec(ctx, sql,
		logEntry.UID, logEntry.User, logEntry.Success,
		logEntry.Reason, logEntry.IP, logEntry.UserAgent)
	return err
}

// ---------- 封号表 ----------

// GetActiveBan 查询用户当前有效的封禁记录（expired_at IS NULL 或 > NOW()），
// 若存在多条则只返回最新一条。
func (d *UserDao) GetActiveBan(ctx context.Context, uid int64) (*model.BanRecord, error) {
	sql := `SELECT id, uid, reason, created_at, expired_at
	        FROM ban_list
	        WHERE uid = $1
	          AND (expired_at IS NULL OR expired_at > NOW())
	        ORDER BY created_at DESC
	        LIMIT 1`
	row := d.db.QueryRow(ctx, sql, uid)
	ban := &model.BanRecord{}
	err := row.Scan(&ban.ID, &ban.UID, &ban.Reason, &ban.CreatedAt, &ban.ExpiredAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return ban, nil
}

// InsertBan 新增封禁记录。
func (d *UserDao) InsertBan(ctx context.Context, ban *model.BanRecord) error {
	sql := `INSERT INTO ban_list (uid, reason, created_at, expired_at)
	        VALUES ($1, $2, NOW(), $3)`
	_, err := d.db.Exec(ctx, sql, ban.UID, ban.Reason, ban.ExpiredAt)
	return err
}

// ---------- 用户设备表 ----------

// GetUserDevices 获取用户的所有设备信息。
func (d *UserDao) GetUserDevices(ctx context.Context, uid int64) ([]model.UserDevice, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, uid, device_id, device_name, platform, push_token, last_active, created_at
		 FROM user_device WHERE uid = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []model.UserDevice
	for rows.Next() {
		var dev model.UserDevice
		err := rows.Scan(&dev.ID, &dev.UID, &dev.DeviceID, &dev.DeviceName,
			&dev.Platform, &dev.PushToken, &dev.LastActive, &dev.CreatedAt)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}
	return devices, rows.Err()
}

// UpsertUserDevice 插入或更新用户设备（基于 uid + device_id 唯一约束）。
func (d *UserDao) UpsertUserDevice(ctx context.Context, dev *model.UserDevice) error {
	sql := `INSERT INTO user_device 
	        (uid, device_id, device_name, platform, push_token, last_active, created_at)
	        VALUES ($1, $2, $3, $4, $5, $6, NOW())
	        ON CONFLICT (uid, device_id) DO UPDATE SET
	            device_name = EXCLUDED.device_name,
	            platform = EXCLUDED.platform,
	            push_token = EXCLUDED.push_token,
	            last_active = EXCLUDED.last_active`
	_, err := d.db.Exec(ctx, sql,
		dev.UID, dev.DeviceID, dev.DeviceName, dev.Platform,
		dev.PushToken, dev.LastActive)
	return err
}
