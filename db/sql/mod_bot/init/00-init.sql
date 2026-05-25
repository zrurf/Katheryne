-- Katheryne Bot Service Database Schema

CREATE TYPE "bot_status" AS ENUM ('ACTIVE', 'DISABLED', 'DELETED');
CREATE TYPE "bot_template_category" AS ENUM ('OFFICIAL', 'COMMUNITY', 'CUSTOM');
CREATE TYPE "bot_template_status" AS ENUM ('DRAFT', 'PUBLISHED', 'DEPRECATED', 'DELETED');
CREATE TYPE "bot_instance_status" AS ENUM ('ACTIVE', 'PAUSED', 'DELETED');

-- ============================================================
-- Bot 身份表（OAuth2 登录实体，保持不变）
-- ============================================================
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

-- ============================================================
-- Bot 模板（开发者创建的蓝图）
-- ============================================================
CREATE TABLE IF NOT EXISTS "bot_template" (
    "id" BIGSERIAL PRIMARY KEY,
    "template_id" BIGINT NOT NULL UNIQUE,             -- 业务 ID
    "name" VARCHAR(64) NOT NULL,
    "avatar" TEXT,
    "description" TEXT,
    "owner_uid" BIGINT NOT NULL,                      -- 模板开发者 UID
    "category" "bot_template_category" NOT NULL DEFAULT 'COMMUNITY',
    "version" VARCHAR(16) NOT NULL DEFAULT '1.0',
    "system_prompt" TEXT NOT NULL,                    -- 系统提示词（模板）
    "welcome_message" TEXT,                           -- 欢迎语
    "conversation_style" JSONB NOT NULL DEFAULT '{}', -- 对话风格配置
    -- {
    --   "multi_message": true,         -- 是否分多条回复
    --   "split_delimiter": "[SPLIT]",  -- 分条分隔符
    --   "use_emoji": true,             -- 是否使用表情
    --   "reply_tone": "friendly",      -- 回复语气
    --   "max_messages_per_turn": 5     -- 单次最多回复条数
    -- }
    "tool_definitions" JSONB NOT NULL DEFAULT '[]',  -- 工具定义列表
    "kb_structure" JSONB NOT NULL DEFAULT '{}',       -- 知识库结构定义（不含具体数据）
    -- {
    --   "require_kb": false,           -- 是否必须配置知识库
    --   "kb_description": "需要上传项目文档"
    -- }
    "config_schema" JSONB NOT NULL DEFAULT '{}',      -- 实例化时需要填写的配置项 schema
    -- {
    --   "api_key": {"type": "string", "required": true, "label": "LLM API Key"},
    --   "model": {"type": "select", "required": true, "options": ["gpt-4", "gpt-3.5"]},
    --   "kb_source": {"type": "kb_picker", "required": false, "label": "知识库"}
    -- }
    "supported_models" JSONB NOT NULL DEFAULT '[]',   -- 支持的模型列表
    "is_official" BOOLEAN NOT NULL DEFAULT FALSE,     -- 是否官方模板
    "display_order" INT NOT NULL DEFAULT 0,           -- 社区展示排序
    "tags" TEXT[] NOT NULL DEFAULT '{}',              -- 标签（搜索用）
    "status" "bot_template_status" NOT NULL DEFAULT 'DRAFT',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_bt_owner" ON "bot_template"("owner_uid", "status");
CREATE INDEX "idx_bt_category" ON "bot_template"("category", "status");
CREATE INDEX "idx_bt_official" ON "bot_template"("is_official", "status");

-- ============================================================
-- Bot 实例（用户从模板创建的部署）
-- ============================================================
CREATE TABLE IF NOT EXISTS "bot_instance" (
    "id" BIGSERIAL PRIMARY KEY,
    "instance_id" BIGINT NOT NULL UNIQUE,             -- 业务 ID
    "bot_id" BIGINT NOT NULL UNIQUE,                  -- 关联的 Bot 身份（OAuth2）
    "template_id" BIGINT NOT NULL,                    -- 引用的模板
    "owner_uid" BIGINT NOT NULL,                      -- 实例拥有者（消费者）
    "name" VARCHAR(64) NOT NULL,
    "avatar" TEXT,
    "is_self_hosted" BOOLEAN NOT NULL DEFAULT FALSE,  -- true=自托管 false=他人托管
    "hosted_by" BIGINT DEFAULT NULL,                  -- 托管者 UID（NULL=自托管，0=官方，其他=第三方）
    "model_provider" VARCHAR(32),                     -- openai / anthropic 等
    "model_name" VARCHAR(64),                         -- gpt-4 / claude-3 等
    "api_key_encrypted" TEXT,                         -- 加密的 API Key（自托管时填写）
    "api_base_url" TEXT,                              -- 自定义 API 地址
    "kb_config" JSONB NOT NULL DEFAULT '{}',          -- 知识库配置
    -- {
    --   "source": "platform",           -- platform / feishu / notion / external
    --   "kb_ids": ["kb_xxx"],           -- 平台知识库 ID 列表
    --   "external_kb": {                -- 外部知识库配置
    --     "type": "feishu",
    --     "url": "https://xxx.feishu.cn/...",
    --     "auth": {"app_id": "...", "app_secret": "..."},
    --     "sync_interval": 3600         -- 同步间隔（秒）
    --   }
    -- }
    "instance_config" JSONB NOT NULL DEFAULT '{}',    -- 实例自定义配置
    "bot_token" VARCHAR(255),                         -- Bot 连接 ws-gateway 的 token
    "status" "bot_instance_status" NOT NULL DEFAULT 'ACTIVE',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_bi_owner" ON "bot_instance"("owner_uid", "status");
CREATE INDEX "idx_bi_template" ON "bot_instance"("template_id", "status");
CREATE INDEX "idx_bi_hosted" ON "bot_instance"("hosted_by", "status");
CREATE INDEX "idx_bi_self_hosted" ON "bot_instance"("is_self_hosted", "status");

-- ============================================================
-- 原有表保持不变
-- ============================================================

-- Bot 凭证
CREATE TABLE IF NOT EXISTS "bot_credential" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL UNIQUE,
    "client_id" VARCHAR(64) NOT NULL UNIQUE,
    "client_secret" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Bot 安装记录
CREATE TABLE IF NOT EXISTS "bot_installation" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,
    "conv_type" VARCHAR(16) NOT NULL DEFAULT 'DM',
    "group_id" BIGINT,
    "installed_by" BIGINT NOT NULL,
    "permissions" TEXT NOT NULL DEFAULT '[]',
    "status" VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    "installed_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP,
    "removed_at" TIMESTAMP,
    UNIQUE ("bot_id", "conv_id")
);

-- 事件投递日志
CREATE TABLE IF NOT EXISTS "bot_event_delivery" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,
    "event_type" VARCHAR(64) NOT NULL,
    "event_id" VARCHAR(128) NOT NULL,
    "payload" JSONB NOT NULL,
    "delivery_method" VARCHAR(16) NOT NULL,
    "status" VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    "retry_count" INT NOT NULL DEFAULT 0,
    "max_retries" INT NOT NULL DEFAULT 5,
    "next_retry_at" TIMESTAMP,
    "last_error" TEXT,
    "delivered_at" TIMESTAMP,
    "response_code" INT,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE ("bot_id", "event_id")
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

-- OAuth2 授权码
CREATE TABLE IF NOT EXISTS "bot_auth_code" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" VARCHAR(128) NOT NULL UNIQUE,
    "client_id" VARCHAR(64) NOT NULL,
    "redirect_uri" TEXT NOT NULL,
    "scope" VARCHAR(256) NOT NULL DEFAULT 'message.read',
    "uid" BIGINT NOT NULL,
    "conv_id" BIGINT,
    "used" BOOLEAN NOT NULL DEFAULT FALSE,
    "expires_at" TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- OAuth2 Token 记录
CREATE TABLE IF NOT EXISTS "bot_oauth_token" (
    "id" BIGSERIAL PRIMARY KEY,
    "bot_id" BIGINT NOT NULL,
    "client_id" VARCHAR(64) NOT NULL,
    "access_token_hash" VARCHAR(128) NOT NULL UNIQUE,
    "refresh_token_hash" VARCHAR(128),
    "scope" VARCHAR(256) NOT NULL DEFAULT 'message.read',
    "revoked" BOOLEAN NOT NULL DEFAULT FALSE,
    "expires_at" TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_bot_inst_conv" ON "bot_installation" ("conv_id", "status")
    WHERE "status" = 'ACTIVE';
CREATE INDEX IF NOT EXISTS "idx_bot_inst_bot" ON "bot_installation" ("bot_id", "status")
    WHERE "status" = 'ACTIVE';
CREATE INDEX IF NOT EXISTS "idx_bot_auth_code" ON "bot_auth_code" ("code", "used", "expires_at")
    WHERE "used" = FALSE;
CREATE INDEX IF NOT EXISTS "idx_bot_oauth_token" ON "bot_oauth_token" ("bot_id", "revoked")
    WHERE "revoked" = FALSE;
CREATE INDEX IF NOT EXISTS "idx_bot_event_retry" ON "bot_event_delivery" ("status", "next_retry_at")
    WHERE "status" IN ('PENDING', 'FAILED') AND "retry_count" < 5;
CREATE INDEX IF NOT EXISTS "idx_bot_event_conv" ON "bot_event_delivery" ("conv_id", "created_at" DESC);