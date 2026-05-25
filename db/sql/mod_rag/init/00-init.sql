-- Katheryne RAG Service Database Schema

-- 知识库（桶）
CREATE TABLE IF NOT EXISTS "kb" (
    "kb_id" VARCHAR(64) PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT DEFAULT '',
    "owner_uid" BIGINT NOT NULL,
    "source_type" VARCHAR(32) NOT NULL DEFAULT 'PLATFORM',  -- PLATFORM / FEISHU / NOTION / WEB
    "source_config" JSONB NOT NULL DEFAULT '{}',            -- 外部数据源配置
    "status" VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    "doc_count" BIGINT NOT NULL DEFAULT 0,
    "chunk_count" BIGINT NOT NULL DEFAULT 0,
    "total_size" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_kb_owner" ON "kb"("owner_uid");
CREATE INDEX "idx_kb_status" ON "kb"("status");
CREATE INDEX "idx_kb_source" ON "kb"("source_type");

-- 外部知识库同步记录
CREATE TABLE IF NOT EXISTS "kb_external_sync" (
    "sync_id" VARCHAR(64) PRIMARY KEY,
    "kb_id" VARCHAR(64) NOT NULL,
    "source_type" VARCHAR(32) NOT NULL,         -- FEISHU / NOTION / WEB
    "source_config" JSONB NOT NULL DEFAULT '{}', -- API Key, URL 等
    "last_synced_at" TIMESTAMP,
    "sync_status" VARCHAR(16) NOT NULL DEFAULT 'PENDING',  -- PENDING / SYNCING / SYNCED / FAILED
    "sync_error" TEXT DEFAULT '',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_kb_ext_sync_kb" ON "kb_external_sync"("kb_id");
CREATE INDEX "idx_kb_ext_sync_status" ON "kb_external_sync"("sync_status");

-- 知识库文档
CREATE TABLE IF NOT EXISTS "kb_document" (
    "doc_id" VARCHAR(64) PRIMARY KEY,
    "kb_id" VARCHAR(64) NOT NULL,
    "file_name" VARCHAR(512) NOT NULL,
    "content_type" VARCHAR(128) DEFAULT '',
    "file_size" BIGINT NOT NULL DEFAULT 0,
    "oss_index" VARCHAR(255) DEFAULT '',
    "status" VARCHAR(16) NOT NULL DEFAULT 'PROCESSING',  -- PROCESSING / READY / FAILED / DELETED
    "chunk_count" INT NOT NULL DEFAULT 0,
    "error_msg" TEXT DEFAULT '',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_kb_doc_kb" ON "kb_document"("kb_id");
CREATE INDEX "idx_kb_doc_status" ON "kb_document"("status");

-- 文档分块
CREATE TABLE IF NOT EXISTS "kb_chunk" (
    "chunk_id" VARCHAR(128) PRIMARY KEY,
    "doc_id" VARCHAR(64) NOT NULL,
    "kb_id" VARCHAR(64) NOT NULL,
    "content" TEXT NOT NULL,
    "chunk_index" INT NOT NULL DEFAULT 0,
    "token_count" BIGINT NOT NULL DEFAULT 0,
    "entities" TEXT DEFAULT '[]',  -- JSON array of entity strings
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX "idx_kb_chunk_doc" ON "kb_chunk"("doc_id");
CREATE INDEX "idx_kb_chunk_kb" ON "kb_chunk"("kb_id");

-- Bot 知识库访问授权（一级：Bot 级别）
CREATE TABLE IF NOT EXISTS "kb_bot_access" (
    "uid" BIGINT NOT NULL,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,
    "granted_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("uid", "bot_id", "conv_id")
);

-- 知识库细粒度授权（二级：知识库级别）
CREATE TABLE IF NOT EXISTS "kb_auth" (
    "kb_id" VARCHAR(64) NOT NULL,
    "bot_id" BIGINT NOT NULL,
    "conv_id" BIGINT NOT NULL,
    "permission" VARCHAR(16) NOT NULL DEFAULT 'READ',  -- READ / READ_WRITE
    "granted_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("kb_id", "bot_id", "conv_id")
);
CREATE INDEX "idx_kb_auth_bot" ON "kb_auth"("bot_id", "conv_id");