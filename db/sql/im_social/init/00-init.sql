-- 群组状态
CREATE TYPE "group_status" AS ENUM (
    'ACTIVE', 'FROZEN', 'SUSPENDED'
);

-- 群成员角色
CREATE TYPE "group_role" AS ENUM (
    'OWNER', 'ADMIN', 'MEMBER'
);

-- 用户关系
CREATE TABLE IF NOT EXISTS "friendship" (
    "uid"        BIGINT NOT NULL,        -- 用户ID
    "peer_uid"   BIGINT NOT NULL,        -- 好友UID
    "remark"     VARCHAR(64),            -- 备注
    "group_name" VARCHAR(64) DEFAULT '', -- 分组名称
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(), -- 创建时间
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(), -- 更新时间
    PRIMARY KEY ("uid", "peer_uid"),
    CHECK ("uid" < "peer_uid")
);

-- 黑名单
CREATE TABLE IF NOT EXISTS "blacklist" (
    "uid"        BIGINT NOT NULL, -- 用户ID
    "peer_uid"   BIGINT NOT NULL, -- 对方UID
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("uid", "peer_uid")
);

-- 会话
CREATE TABLE IF NOT EXISTS "conversations" (
    "conv_id" BIGSERIAL PRIMARY KEY,
    "type" VARCHAR(16) NOT NULL,                      -- SINGLE / GROUP
    "group_id" BIGINT,                                -- 群聊关联的 group_id（单聊为 NULL）
    "uid" BIGINT,                                     -- 单聊用户A
    "peer_uid" BIGINT,                                -- 单聊用户B
    "name" VARCHAR(64),                               -- 会话名称
    "avatar" TEXT,                                    -- 会话头像缓存
    "last_msg_id" BIGINT,                             -- 最后一条消息 ID
    "last_msg_snippet" TEXT,                          -- 最后消息摘要
    "last_msg_time" TIMESTAMP,                        -- 最后消息时间
    "last_msg_sender" BIGINT,                         -- 最后消息发送者
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT "conv_type_check" CHECK ("type" IN ('SINGLE', 'GROUP')),
    CONSTRAINT "conv_single_check" CHECK (
        ("type" = 'SINGLE' AND "uid" IS NOT NULL AND "peer_uid" IS NOT NULL AND "group_id" IS NULL)
        OR
        ("type" = 'GROUP' AND "group_id" IS NOT NULL AND "uid" IS NULL AND "peer_uid" IS NULL)
    ),
    CHECK ("uid" < "peer_uid")
);

-- 单聊去重：同一对用户只有一个会话
CREATE UNIQUE INDEX IF NOT EXISTS "idx_conv_single_pair" ON "conversations" ("uid", "peer_uid")
    WHERE "type" = 'SINGLE';
-- 群聊通过 group_id 关联
CREATE UNIQUE INDEX IF NOT EXISTS "idx_conv_group_id" ON "conversations" ("group_id")
    WHERE "type" = 'GROUP';

-- 会话成员表（统一单聊/群聊的成员视图）
CREATE TABLE IF NOT EXISTS "conv_members" (
    "conv_id" BIGINT NOT NULL,
    "uid" BIGINT NOT NULL,
    "mute" BOOLEAN NOT NULL DEFAULT FALSE,            -- 免打扰
    "pinned" BOOLEAN NOT NULL DEFAULT FALSE,          -- 置顶
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,        -- 是否在会话中（退出群聊置 false）
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("conv_id", "uid")
);

CREATE INDEX IF NOT EXISTS "idx_conv_members_uid" ON "conv_members" ("uid") WHERE "is_active" = TRUE;

-- 群组
CREATE TABLE IF NOT EXISTS "groups" (
    "id"                        BIGSERIAL PRIMARY KEY,  -- 主键ID
    "group_id"                  BIGINT NOT NULL UNIQUE, -- 群组ID
    "name"                      VARCHAR(64) NOT NULL,   -- 群组名称
    "avatar"                    TEXT,                   -- 群组头像URL
    "owner"                     BIGINT NOT NULL,        -- 群组创建者ID
    "member_count"              INT DEFAULT 1,          -- 成员数量
    "status" "group_status"     NOT NULL DEFAULT 'ACTIVE', -- 状态
    "verify_mode"               VARCHAR(64) NOT NULL DEFAULT 'NONE', -- 成员验证模式"
    "created_at"                TIMESTAMP NOT NULL DEFAULT NOW(), -- 创建时间

    CONSTRAINT "verify_mode_check" CHECK ("verify_mode" IN ('NONE', 'ADMIN_CONFIRM'))
);

-- 群成员
CREATE TABLE IF NOT EXISTS "group_members" (
    "group_id"              BIGINT NOT NULL,           -- 群组ID
    "uid"                   BIGINT NOT NULL,           -- 用户ID
    "role" "group_role"     NOT NULL DEFAULT 'MEMBER', -- 群角色
    "nick"                  VARCHAR(64),               -- 群昵称
    "join_time"             TIMESTAMP NOT NULL DEFAULT NOW(), -- 加群时间
    "inviter"               BIGINT,                    -- 邀请人ID（主动加群为NULL）
    "mute_until"            TIMESTAMP,                 -- 禁言到期时间（未禁言为NULL）

    PRIMARY KEY ("group_id", "uid")
);

-- 群公告
CREATE TABLE IF NOT EXISTS "group_announcement" (
    "id" BIGSERIAL PRIMARY KEY,
    "group_id" BIGINT NOT NULL,
    "uid" BIGINT NOT NULL, -- 发布者
    "content" TEXT NOT NULL,
    "pinned" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_friendship_uid" ON "friendship" ("uid");
CREATE INDEX IF NOT EXISTS "idx_friendship_peer_uid" ON "friendship" ("peer_uid");
CREATE INDEX IF NOT EXISTS "idx_blacklist_uid" ON "blacklist" ("uid");
CREATE INDEX IF NOT EXISTS "idx_blacklist_peer_uid" ON "blacklist" ("peer_uid");
CREATE INDEX IF NOT EXISTS "idx_groups_group_id" ON "groups" ("group_id");
CREATE INDEX IF NOT EXISTS "idx_groups_owner" ON "groups" ("owner");
CREATE INDEX IF NOT EXISTS "idx_conversations_last_msg_time" ON "conversations" ("last_msg_time" DESC);
