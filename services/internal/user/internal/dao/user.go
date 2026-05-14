package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User 对应 users 表
type User struct {
	UID          int64
	Name         string
	Phone        string
	Avatar       *string
	Status       string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	LastLogin    time.Time
	OpaqueRecord []byte
}

// UserConfig 对应 user_config 表
type UserConfig struct {
	UID             int64
	Language        string
	MsgNotification bool
	SoundEnabled    bool
	AutoTranslate   bool
	TranslateTarget *string
	ContentFilter   bool
	Enable2FA       bool
	TOTPSecret      *string
	TOTPBackupCodes []string
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

// BanRecord 对应 ban_list 表
type BanRecord struct {
	ID        int64
	UID       int64
	Reason    string
	CreatedAt time.Time
	ExpiredAt *time.Time
}

// UserDevice 对应 user_device 表
type UserDevice struct {
	ID         int64
	UID        int64
	DeviceID   string
	DeviceName *string
	Platform   *string
	PushToken  *string
	LastActive *time.Time
	CreatedAt  time.Time
}

// UserDao 用户数据访问层
type UserDao struct {
	db *pgxpool.Pool
}

func NewUserDao(pool *pgxpool.Pool) *UserDao {
	return &UserDao{db: pool}
}

// ---------- 基础用户操作 ----------

// CreateUser 创建用户
func (d *UserDao) CreateUser(ctx context.Context, uid int64, phone, name string, record []byte) (*User, error) {
	sql := `INSERT INTO users (uid, phone, name, opaque_record, status, created_at, updated_at, last_login)
	        VALUES ($1, $2, $3, $4, 'ACTIVE', NOW(), NOW(), NOW())
	        RETURNING uid, name, phone, avatar, status, created_at, updated_at, last_login`
	u := &User{OpaqueRecord: record}
	err := d.db.QueryRow(ctx, sql, uid, phone, name, record).Scan(
		&u.UID, &u.Name, &u.Phone, &u.Avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUID 根据 UID 获取用户信息
func (d *UserDao) GetUserByUID(ctx context.Context, uid int64) (*User, error) {
	row := d.db.QueryRow(ctx,
		`SELECT uid, name, phone, avatar, status, created_at, updated_at, last_login
		 FROM users WHERE uid = $1`, uid)
	u := &User{}
	var avatar *string
	err := row.Scan(&u.UID, &u.Name, &u.Phone, &avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
	if err != nil {
		return nil, err
	}
	u.Avatar = avatar
	return u, nil
}

// GetUsersByUIDs 批量获取用户信息
func (d *UserDao) GetUsersByUIDs(ctx context.Context, uids []int64) ([]*User, error) {
	if len(uids) == 0 {
		return []*User{}, nil
	}
	placeholders := make([]string, len(uids))
	args := make([]interface{}, len(uids))
	for i, uid := range uids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = uid
	}
	query := fmt.Sprintf(
		`SELECT uid, name, phone, avatar, status, created_at, updated_at, last_login
		 FROM users WHERE uid IN (%s)`, strings.Join(placeholders, ","))
	rows, err := d.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		var avatar *string
		err := rows.Scan(&u.UID, &u.Name, &u.Phone, &avatar, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
		if err != nil {
			return nil, err
		}
		u.Avatar = avatar
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUser 更新用户基本信息
func (d *UserDao) UpdateUser(ctx context.Context, uid int64, name, avatar string) error {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}
	if avatar != "" {
		setClauses = append(setClauses, fmt.Sprintf("avatar = $%d", argIdx))
		args = append(args, avatar)
		argIdx++
	}
	if len(setClauses) == 0 {
		return nil
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE users SET %s WHERE uid = $%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, uid)
	_, err := d.db.Exec(ctx, query, args...)
	return err
}

// UpdateUserStatus 更新用户状态
func (d *UserDao) UpdateUserStatus(ctx context.Context, uid int64, status string) error {
	sql := `UPDATE users SET status = $1, updated_at = NOW() WHERE uid = $2`
	_, err := d.db.Exec(ctx, sql, status, uid)
	return err
}

// UpdateLastLogin 更新最后登录时间
func (d *UserDao) UpdateLastLogin(ctx context.Context, uid int64) error {
	sql := `UPDATE users SET last_login = NOW() WHERE uid = $1`
	_, err := d.db.Exec(ctx, sql, uid)
	return err
}

// PhoneExists 检查手机号是否已注册
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

// ---------- 用户配置表 ----------

// GetUserConfig 获取用户配置
func (d *UserDao) GetUserConfig(ctx context.Context, uid int64) (*UserConfig, error) {
	cfg := &UserConfig{}
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

// UpsertUserConfig 插入或更新用户配置
func (d *UserDao) UpsertUserConfig(ctx context.Context, cfg *UserConfig) error {
	// 确保配置行存在
	const insertDefault = `INSERT INTO user_config (uid, language, msg_notification, sound_enabled, auto_translate,
		translate_target, content_filter, enable_2fa, created_at, updated_at)
		VALUES ($1, 'zh-CN', true, true, false, null, true, false, NOW(), NOW())
		ON CONFLICT (uid) DO NOTHING`
	_, err := d.db.Exec(ctx, insertDefault, cfg.UID)
	if err != nil {
		return err
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	add := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	add("language", cfg.Language)
	add("msg_notification", cfg.MsgNotification)
	add("sound_enabled", cfg.SoundEnabled)
	add("auto_translate", cfg.AutoTranslate)
	if cfg.TranslateTarget != nil {
		add("translate_target", *cfg.TranslateTarget)
	} else {
		add("translate_target", nil)
	}
	add("content_filter", cfg.ContentFilter)
	add("enable_2fa", cfg.Enable2FA)
	if cfg.TOTPSecret != nil {
		add("totp_secret", *cfg.TOTPSecret)
	} else {
		add("totp_secret", nil)
	}
	add("totp_backup_codes", cfg.TOTPBackupCodes)

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE user_config SET %s WHERE uid = $%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, cfg.UID)
	_, err = d.db.Exec(ctx, query, args...)
	return err
}

// ---------- 封号表 ----------

// GetActiveBan 查询用户当前有效的封禁记录
func (d *UserDao) GetActiveBan(ctx context.Context, uid int64) (*BanRecord, error) {
	sql := `SELECT id, uid, reason, created_at, expired_at
	        FROM ban_list
	        WHERE uid = $1
	          AND (expired_at IS NULL OR expired_at > NOW())
	        ORDER BY created_at DESC
	        LIMIT 1`
	row := d.db.QueryRow(ctx, sql, uid)
	ban := &BanRecord{}
	err := row.Scan(&ban.ID, &ban.UID, &ban.Reason, &ban.CreatedAt, &ban.ExpiredAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ban, nil
}

// ---------- 用户设备表 ----------

// UpsertUserDevice 插入或更新用户设备
func (d *UserDao) UpsertUserDevice(ctx context.Context, dev *UserDevice) error {
	sql := `INSERT INTO user_device 
	        (uid, device_id, device_name, platform, push_token, last_active, created_at)
	        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	        ON CONFLICT (uid, device_id) DO UPDATE SET
	            device_name = EXCLUDED.device_name,
	            platform = EXCLUDED.platform,
	            push_token = EXCLUDED.push_token,
	            last_active = EXCLUDED.last_active`
	_, err := d.db.Exec(ctx, sql,
		dev.UID, dev.DeviceID, dev.DeviceName, dev.Platform,
		dev.PushToken)
	return err
}

// RemoveUserDevice 删除用户设备
func (d *UserDao) RemoveUserDevice(ctx context.Context, uid int64, deviceID string) error {
	sql := `DELETE FROM user_device WHERE uid = $1 AND device_id = $2`
	_, err := d.db.Exec(ctx, sql, uid, deviceID)
	return err
}

// GetUserDevices 获取用户的所有设备
func (d *UserDao) GetUserDevices(ctx context.Context, uid int64) ([]*UserDevice, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, uid, device_id, device_name, platform, push_token, last_active, created_at
		 FROM user_device WHERE uid = $1`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*UserDevice
	for rows.Next() {
		dev := &UserDevice{}
		err := rows.Scan(&dev.ID, &dev.UID, &dev.DeviceID, &dev.DeviceName,
			&dev.Platform, &dev.PushToken, &dev.LastActive, &dev.CreatedAt)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}
	return devices, rows.Err()
}
