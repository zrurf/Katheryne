-- 用户消息
CREATE TABLE IF NOT EXISTS "message" (
    "id"            BIGSERIAL PRIMARY KEY,   -- 消息ID
    "conv_id"       BIGINT NOT NULL,         -- 会话ID
    "sender"        BIGINT NOT NULL,         -- 发送者ID
    "receiver"      BIGINT NOT NULL,         -- 接收者ID（群消息为群ID）
    "type"          VARCHAR(16) NOT NULL,    -- 消息类型
    "content"       TEXT NOT NULL,           -- 内容（如果非text则为对象引用）
    "content_type"  VARCHAR(255) NOT NULL DEFAULT 'text', -- 内容类型
    "quote_msg_id"  BIGINT,                  -- 引用消息ID
    "recalled"      BOOLEAN NOT NULL DEFAULT FALSE, -- 是否被撤回
    "recall_time"   TIMESTAMP,               -- 撤回时间
    "edited"        BOOLEAN NOT NULL DEFAULT FALSE, -- 是否已编辑
    "edited_at"     TIMESTAMP,               -- 最后编辑时间
    "edit_count"    INT NOT NULL DEFAULT 0,  -- 编辑次数
    "extra"         JSONB,                   -- 额外信息
    "created_at"    TIMESTAMP DEFAULT NOW(), -- 时间戳
    CONSTRAINT "msg_type_check" CHECK ("type" IN ('text', 'image', 'file', 'voice', 'video', 'system', 'card'))
);

-- 已读区间表
CREATE TABLE IF NOT EXISTS "msg_read_intervals" (
    "id"            BIGSERIAL PRIMARY KEY,
    "conv_id"       BIGINT NOT NULL,        -- 会话ID (单聊或群聊ID)
    "uid"           BIGINT NOT NULL,        -- 阅读者UID
    "start_msg_id"  BIGINT NOT NULL,        -- 区间起始ID (闭区间)
    "end_msg_id"    BIGINT NOT NULL,        -- 区间结束ID (闭区间)
    "created_at"    TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE ("conv_id", "uid", "start_msg_id")
);

-- 消息索引（核心查询路径）
CREATE INDEX IF NOT EXISTS "message_conv_id_idx" ON "message"("conv_id");
-- 会话内按 ID 正序翻页（cursor-based pagination）
CREATE INDEX IF NOT EXISTS "message_conv_id_id_idx" ON "message"("conv_id", "id");
-- 会话内按时间倒序（首屏加载）
CREATE INDEX IF NOT EXISTS "message_conv_id_time_idx" ON "message"("conv_id", "created_at" DESC);
-- 发送者维度的消息查询
CREATE INDEX IF NOT EXISTS "message_sender_idx" ON "message"("sender");
-- 全局时间索引（管理员审计、漫游同步）
CREATE INDEX IF NOT EXISTS "message_created_at_idx" ON "message"("created_at");
-- 被引用的消息查询
CREATE INDEX IF NOT EXISTS "message_quote_idx" ON "message"("quote_msg_id") WHERE "quote_msg_id" IS NOT NULL;
-- 关键词搜索（全文检索，simple 词干分析器支持多语言基础搜索）
CREATE INDEX IF NOT EXISTS "message_content_fts_idx" ON "message" USING GIN (to_tsvector('simple', "content"))
    WHERE "type" = 'text' AND "recalled" = FALSE;

CREATE INDEX IF NOT EXISTS "idx_interval_lookup" ON "msg_read_intervals" ("conv_id", "end_msg_id") INCLUDE ("uid", "start_msg_id");
CREATE INDEX IF NOT EXISTS "idx_interval_user" ON "msg_read_intervals" ("conv_id", "uid");