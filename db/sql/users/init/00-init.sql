-- 用户表
CREATE TABLE IF NOT EXISTS "users" (
  "uid"             BIGINT PRIMARY KEY,                 -- 用户ID
  "name"            VARCHAR(64) NOT NULL,               -- 用户名
  "phone"           VARCHAR(32) NOT NULL UNIQUE,        -- 手机号
  "avatar"          TEXT,                               -- 头像
  "status"          VARCHAR(64) NOT NULL DEFAULT 'ACTIVE', -- 状态
  "updated_at"      TIMESTAMP NOT NULL DEFAULT NOW(),   -- 更新时间
  "created_at"      TIMESTAMP NOT NULL DEFAULT NOW(),   -- 创建时间
  "last_login"      TIMESTAMP NOT NULL DEFAULT NOW(),   -- 最后登录时间
  "opaque_record"   BYTEA NOT NULL,                     -- OPAQUE注册记录

  CONSTRAINT "user_status_check" CHECK (status IN ('ACTIVE', 'INACTIVE', 'FROZEN', 'BANNED', 'ARCHIVED'))
);

-- 用户配置表
CREATE TABLE IF NOT EXISTS "user_config" (
    "uid"                 BIGINT PRIMARY KEY REFERENCES "users"("uid") ON DELETE CASCADE,
    "language"            VARCHAR(10) NOT NULL DEFAULT 'zh-CN',
    "msg_notification"    BOOLEAN NOT NULL DEFAULT TRUE,
    "sound_enabled"       BOOLEAN NOT NULL DEFAULT TRUE,
    "auto_translate"      BOOLEAN NOT NULL DEFAULT FALSE,
    "translate_target"    VARCHAR(10),
    "content_filter"      BOOLEAN NOT NULL DEFAULT TRUE,
    "enable_2fa"          BOOLEAN NOT NULL DEFAULT FALSE,   -- 是否启用2FA
    "totp_secret"         TEXT,                             -- TOTP 密钥
    "totp_backup_codes"   TEXT[],                           -- TOTP备用代码
    "updated_at"          TIMESTAMP NOT NULL DEFAULT NOW(),
    "created_at"          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 封号表
CREATE TABLE IF NOT EXISTS "ban_list" (
  "id"         BIGSERIAL PRIMARY KEY, -- 封禁ID
  "uid"        BIGINT NOT NULL,       -- 用户ID
  "reason"     TEXT NOT NULL,         -- 封禁理由
  "created_at" TIMESTAMP NOT NULL DEFAULT NOW(), -- 封禁时间
  "expired_at" TIMESTAMP,             -- 封禁有效期（永久封禁为NULL）

  CONSTRAINT "ban_list_expired_gt_created" CHECK ("expired_at" IS NULL OR "expired_at" > "created_at")
);

-- 登录日志
CREATE TABLE IF NOT EXISTS "login_log" (
  "id"         BIGSERIAL PRIMARY KEY,  -- 记录ID
  "uid"        BIGINT REFERENCES "users"("uid") ON DELETE SET NULL, -- UID
  "user"       TEXT NOT NULL,          -- 用户名
  "success"    BOOLEAN NOT NULL,       -- 状态
  "reason"     TEXT,                   -- 失败原因
  "ip"         INET,                   -- 登录IP
  "user_agent" TEXT,                   -- UA
  "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 用户设备
CREATE TABLE IF NOT EXISTS "user_device" (
    "id"            BIGSERIAL PRIMARY KEY,
    "uid"           BIGINT NOT NULL REFERENCES "users"("uid") ON DELETE CASCADE,
    "device_id"     VARCHAR(64) NOT NULL,
    "device_name"   VARCHAR(64),
    "platform"      VARCHAR(32),   -- ios, android, web, cli
    "push_token"    TEXT,
    "last_active"   TIMESTAMP DEFAULT NOW(),
    "created_at"    TIMESTAMP DEFAULT NOW(),
    UNIQUE("uid", "device_id")
);

CREATE INDEX IF NOT EXISTS "idx_users_name" ON "users"("name");
CREATE INDEX IF NOT EXISTS "idx_users_phone" ON "users"("phone");
CREATE INDEX IF NOT EXISTS "idx_ban_list_uid" ON "ban_list"("uid");
CREATE INDEX IF NOT EXISTS "idx_login_log_uid" ON "login_log"("uid");