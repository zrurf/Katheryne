-- ============================================================
-- 官方 AI Bot (Katheryne AI) 初始化
-- ============================================================
-- 系统用户 UID = 0 为平台保留的系统账号
-- Bot ID = 10001 为官方 AI Bot 保留 ID
-- ============================================================

-- 官方 AI Bot 注册
INSERT INTO "bot" ("bot_id", "name", "avatar", "description", "owner_uid", "subscribe_events", "status")
VALUES (10001, 'Katheryne AI',
        '/static/avatars/katheryne-ai.png',
        'Katheryne 官方 AI 助手。支持智能问答、消息总结、上下文回复建议、多语言翻译、内容审核等功能。',
        0,
        ARRAY['message.create', 'message.recall', 'message.edit', 'group.join', 'group.leave'],
        'ACTIVE')
ON CONFLICT ("bot_id") DO NOTHING;

-- 官方 AI Bot OAuth2 凭证
-- client_id / client_secret 用于 OAuth2 授权流程
-- 第三方 Bot 也可通过 Bot API 自行注册获得独立凭证
INSERT INTO "bot_credential" ("bot_id", "client_id", "client_secret")
VALUES (10001,
        'katheryne_ai_official_bot',
        'kth_official_secret_v1_production')
ON CONFLICT ("bot_id") DO NOTHING;

-- 官方 AI Bot 限流配额（平台 Bot 配额更高）
INSERT INTO "bot_rate_limit" ("bot_id", "messages_per_minute", "messages_per_day", "api_calls_per_minute")
VALUES (10001, 300, 100000, 600)
ON CONFLICT ("bot_id") DO NOTHING;