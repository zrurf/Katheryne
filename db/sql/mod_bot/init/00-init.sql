CREATE TYPE "bot_status" AS ENUM ('ACTIVE', 'DISABLED', 'DELETED');

-- Bot 定义
CREATE TABLE IF NOT EXISTS "bot" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL UNIQUE,                 -- 业务 ID
    "name" VARCHAR(64) NOT NULL,
    "avatar" TEXT,
    "description" TEXT,
    "owner_uid" BIGINT NOT NULL,                      -- 开发者 UID
    "webhook_url" TEXT,                               -- Webhook 接收地址
    "webhook_secret" VARCHAR(128),                    -- Webhook 签名密钥（创建时生成）
    "subscribe_events" TEXT[] NOT NULL DEFAULT '{}',  -- 订阅的事件列表
    "status" "bot_status" NOT NULL DEFAULT 'ACTIVE',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Bot 凭证
CREATE TABLE IF NOT EXISTS "bot_credential" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL UNIQUE,
    "client_id" VARCHAR(64) NOT NULL UNIQUE,         -- OAuth2 client_id
    "client_secret" VARCHAR(255) NOT NULL,           -- OAuth2 client_secret
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Bot 安装记录
CREATE TABLE IF NOT EXISTS "bot_installation" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,                        -- 安装到哪个会话
    "group_id" BIGINT,                                -- 群聊 ID
    "installed_by" BIGINT NOT NULL,                   -- 谁安装的
    "permissions" TEXT[] NOT NULL DEFAULT '{}',       -- 授权的权限列表
    "status" VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',   -- ACTIVE / REMOVED
    "installed_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "removed_at" TIMESTAMP,
    UNIQUE ("bot_id", "conv_id")
);

-- 事件投递日志
CREATE TABLE IF NOT EXISTS "bot_event_delivery" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,
    "event_type" VARCHAR(64) NOT NULL,                -- 事件类型
    "event_id" VARCHAR(128) NOT NULL,                 -- 幂等 ID（全局唯一）
    "payload" JSONB NOT NULL,                         -- 事件完整内容
    "delivery_method" VARCHAR(16) NOT NULL,           -- websocket / webhook
    "status" VARCHAR(16) NOT NULL DEFAULT 'PENDING',  -- PENDING / DELIVERED / FAILED
    "retry_count" INT NOT NULL DEFAULT 0,
    "max_retries" INT NOT NULL DEFAULT 5,
    "next_retry_at" TIMESTAMP,
    "last_error" TEXT,
    "delivered_at" TIMESTAMP,
    "response_code" INT,                             -- HTTP 响应码（仅 webhook）
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE ("bot_id", "event_id")                    -- 幂等：同一事件不重复投递
);

-- Bot 调用 API 的限流配额
CREATE TABLE IF NOT EXISTS "bot_rate_limit" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL UNIQUE,
    "messages_per_minute" INT NOT NULL DEFAULT 60,
    "messages_per_day" INT NOT NULL DEFAULT 10000,
    "api_calls_per_minute" INT NOT NULL DEFAULT 120,
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_bot_inst_conv" ON "bot_installation" ("conv_id", "status")
    WHERE "status" = 'ACTIVE';
CREATE INDEX IF NOT EXISTS "idx_bot_inst_bot" ON "bot_installation" ("bot_id", "status")
    WHERE "status" = 'ACTIVE';

-- Webhook 重试扫描索引
CREATE INDEX IF NOT EXISTS "idx_bot_event_retry" ON "bot_event_delivery" ("status", "next_retry_at")
    WHERE "status" IN ('PENDING', 'FAILED') AND "retry_count" < 5;
-- 按会话查事件
CREATE INDEX IF NOT EXISTS "idx_bot_event_conv" ON "bot_event_delivery" ("conv_id", "created_at" DESC);