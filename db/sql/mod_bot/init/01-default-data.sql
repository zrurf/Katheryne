-- ============================================================
-- 官方 AI Bot (Katheryne AI) 初始化
-- ============================================================
-- 系统用户 UID = 0 为平台保留的系统账号
-- Bot ID = 10001 为官方 AI Bot 保留 ID
-- ============================================================

-- 官方 Bot 身份
INSERT INTO "bot" ("bot_id", "name", "avatar", "description", "owner_uid", "subscribe_events", "status")
VALUES (10001, 'Katheryne AI',
        '/static/avatars/katheryne-ai.png',
        'Katheryne 官方 AI 助手。支持智能问答、消息总结、上下文回复建议、多语言翻译、内容审核、知识库检索等功能。',
        0,
        ARRAY['message.create', 'message.recall', 'message.edit', 'group.join', 'group.leave'],
        'ACTIVE')
ON CONFLICT ("bot_id") DO NOTHING;

-- 官方 Bot 模板（开发者模板，社区可见）
INSERT INTO "bot_template" (
    "template_id", "name", "avatar", "description", "owner_uid",
    "category", "version", "system_prompt", "welcome_message",
    "conversation_style", "tool_definitions", "kb_structure",
    "config_schema", "supported_models", "is_official",
    "tags", "status"
) VALUES (
    10001,
    'Katheryne AI',
    '/static/avatars/katheryne-ai.png',
    'Katheryne 官方 AI 助手模板。支持智能问答、消息总结、上下文回复建议、多语言翻译、内容审核等功能。可接入本平台知识库或外部知识库。',
    0,
    'OFFICIAL',
    '1.0',
    -- system_prompt
    '你是 Katheryne，Katheryne 即时通讯平台的内置 AI 助手。你的职责是帮助用户解答问题、总结对话、翻译文本、审核内容、提供回复建议。你可以使用 web_search 工具获取实时信息，也可以查询用户授权给你的知识库来回答问题。请保持友好、专业且准确。',
    -- welcome_message
    '你好！我是 Katheryne，你的智能助手。有什么我可以帮助你的吗？',
    -- conversation_style
    '{
        "multi_message": true,
        "split_delimiter": "[SPLIT]",
        "use_emoji": true,
        "reply_tone": "friendly",
        "max_messages_per_turn": 5,
        "max_chars_per_message": 500
    }'::jsonb,
    -- tool_definitions
    '[
        {
            "name": "web_search",
            "description": "搜索互联网获取实时信息（天气、新闻、技术问答等）",
            "parameters": {"query": "string"}
        },
        {
            "name": "kb_search",
            "description": "在已授权的知识库中检索信息",
            "parameters": {"query": "string", "kb_ids": ["string"]}
        },
        {
            "name": "summarize",
            "description": "总结最近的对话内容",
            "parameters": {"conv_id": "int64", "count": "int"}
        },
        {
            "name": "translate",
            "description": "翻译文本",
            "parameters": {"text": "string", "target_lang": "string"}
        }
    ]'::jsonb,
    -- kb_structure
    '{
        "require_kb": false,
        "kb_description": "可配置知识库以增强问答能力。支持本平台知识库或外部知识库（飞书等）。"
    }'::jsonb,
    -- config_schema
    '{
        "api_key": {
            "type": "password",
            "required": true,
            "label": "LLM API Key",
            "description": "大语言模型 API Key（OpenAI / Anthropic 兼容）"
        },
        "model": {
            "type": "select",
            "required": true,
            "label": "模型选择",
            "options": ["gpt-4o", "gpt-4", "gpt-3.5-turbo", "claude-3-opus", "claude-3-sonnet"]
        },
        "api_base_url": {
            "type": "string",
            "required": false,
            "label": "API Base URL",
            "description": "自定义 API 地址（留空使用默认）"
        },
        "kb_source": {
            "type": "kb_picker",
            "required": false,
            "label": "知识库",
            "multiple": true,
            "description": "选择要关联的知识库（可选）"
        }
    }'::jsonb,
    -- supported_models
    '[
        {"provider": "openai", "models": ["gpt-4o", "gpt-4", "gpt-3.5-turbo"]},
        {"provider": "anthropic", "models": ["claude-3-opus", "claude-3-sonnet"]}
    ]'::jsonb,
    TRUE,
    '{"AI助手","官方","智能问答","翻译","总结"}',
    'PUBLISHED'
)
ON CONFLICT ("template_id") DO NOTHING;

-- 官方 Bot 实例（由官方托管，用户直接安装即可使用）
-- bot_id=10001 与 bot 表对应，instance_id=10001 为实例 ID
INSERT INTO "bot_instance" (
    "instance_id", "bot_id", "template_id", "owner_uid",
    "name", "avatar", "is_self_hosted", "hosted_by",
    "model_provider", "model_name",
    "instance_config", "bot_token", "status"
) VALUES (
    10001,          -- instance_id
    10001,          -- bot_id (关联 OAuth2 身份)
    10001,          -- template_id
    0,              -- owner_uid (系统)
    'Katheryne AI',
    '/static/avatars/katheryne-ai.png',
    FALSE,          -- is_self_hosted (官方托管)
    0,              -- hosted_by (官方)
    'openai',       -- model_provider
    'gpt-4o',       -- model_name
    '{"official": true}'::jsonb,
    'kth_official_token_v1',
    'ACTIVE'
)
ON CONFLICT ("instance_id") DO NOTHING;

-- 官方 AI Bot OAuth2 凭证
INSERT INTO "bot_credential" ("bot_id", "client_id", "client_secret")
VALUES (10001,
        'katheryne_ai_official_bot',
        'kth_official_secret_v1_production')
ON CONFLICT ("bot_id") DO NOTHING;

-- 官方 AI Bot 限流配额（平台 Bot 配额更高）
INSERT INTO "bot_rate_limit" ("bot_id", "messages_per_minute", "messages_per_day", "api_calls_per_minute")
VALUES (10001, 300, 100000, 600)
ON CONFLICT ("bot_id") DO NOTHING;