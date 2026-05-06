-- OAuth2 应用
CREATE TABLE IF NOT EXISTS "oauth_app" (
    "id"            BIGSERIAL PRIMARY KEY,       -- 主键ID
    "client_id"     VARCHAR(64) NOT NULL,        -- 客户端ID
    "client_secret" VARCHAR(255) NOT NULL,       -- 密钥
    "client_name"   VARCHAR(64) NOT NULL,        -- 应用名称
    "client_desc"   TEXT NOT NULL,               -- 应用描述
    "redirect_uris" TEXT[] NOT NULL,             -- 允许的重定向地址
    "scopes"        TEXT[] NOT NULL,             -- 允许的权限
    "grant_types"   TEXT[] NOT NULL              -- 允许的授权类型
);

-- 已授权应用
CREATE TABLE IF NOT EXISTS "oauth_authorized_app" (
    "uid" BIGINT NOT NULL,
    "client_id" VARCHAR(64) NOT NULL,
    "scope" TEXT NOT NULL DEFAULT '',
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY ("uid", "client_id")
);

CREATE INDEX IF NOT EXISTS "idx_oauth_app_client_id" ON "oauth_app" ("client_id");
CREATE INDEX IF NOT EXISTS "idx_oauth_authorized_app_uid" ON "oauth_authorized_app" ("uid");