import { createSignal, Show, onMount, For, createResource } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { authStore } from "../stores/auth";
import { appNav } from "../services/nav";
import { themeStore } from "../stores/theme";
import {
  getSavedServers,
  getActiveServer,
  setActiveServerId,
  addServer,
  removeServer,
  generateServerId,
  parseServerUrl,
  getServerApiBase,
  type ServerConfig,
} from "../services/config";
import { checkServerHealth, api } from "../services/api";
import type { BotInfo, CreateBotReq } from "../services/api";
import {
  ArrowLeft,
  Server,
  Plus,
  X,
  Check,
  LogOut,
  User,
  Shield,
  Bell,
  Palette,
  Globe,
  Trash2,
  Bot,
  Edit,
} from "lucide-solid";

export function SettingsPage() {
  const navigate = useNavigate();
  const [servers, setServers] = createSignal<ServerConfig[]>([]);
  const [activeServer, setActiveServer] = createSignal<ServerConfig | null>(null);
  const [newServerUrl, setNewServerUrl] = createSignal("");
  const [addingServer, setAddingServer] = createSignal(false);
  const [serverError, setServerError] = createSignal("");
  const [activeTab, setActiveTab] = createSignal<"general" | "servers" | "account" | "bots" | "community">("general");
  const [editingName, setEditingName] = createSignal(false);
  const [newName, setNewName] = createSignal("");

  const handleAvatarUpload = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      const uploadResp = await api.oss.upload(file);
      const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
      await authStore.updateProfile(undefined, proxyPath);
    } catch {
      alert("头像上传失败，请稍后再试");
    } finally {
      input.value = "";
    }
  };

  const handleSaveName = async () => {
    const name = newName().trim();
    if (!name) return;
    try {
      await authStore.updateProfile(name, undefined);
      setEditingName(false);
    } catch {
      alert("昵称修改失败");
    }
  };

  onMount(() => {
    const saved = getSavedServers();
    setServers(saved);
    const active = getActiveServer();
    setActiveServer(active);
  });

  const handleSelectServer = (server: ServerConfig) => {
    setActiveServerId(server.id);
    setActiveServer(server);
    setServerError("");
  };

  const handleAddServer = async (e: Event) => {
    e.preventDefault();
    if (!newServerUrl().trim()) return;

    setAddingServer(true);
    setServerError("");

    try {
      const { apiUrl, wsUrl, name } = parseServerUrl(newServerUrl());
      const isHealthy = await checkServerHealth(apiUrl);
      if (!isHealthy) {
        setServerError("无法连接到该服务器，请检查地址是否正确");
        setAddingServer(false);
        return;
      }

      const server: ServerConfig = {
        id: generateServerId(),
        name,
        apiUrl,
        wsUrl,
      };
      addServer(server);
      setServers(getSavedServers());
      setActiveServerId(server.id);
      setActiveServer(server);
      setNewServerUrl("");
    } catch (err) {
      setServerError(`服务器配置失败: ${err instanceof Error ? err.message : "未知错误"}`);
    } finally {
      setAddingServer(false);
    }
  };

  const handleRemoveServer = (id: string) => {
    removeServer(id);
    const updated = getSavedServers();
    setServers(updated);
    if (activeServer()?.id === id) {
      setActiveServer(updated.length > 0 ? updated[0] : null);
      if (updated.length > 0) {
        setActiveServerId(updated[0].id);
      }
    }
  };

  const handleLogout = async () => {
    await authStore.logout();
    navigate("/login", { replace: true });
  };

  const [bots, setBots] = createSignal<BotInfo[]>([]);
  const [botsLoading, setBotsLoading] = createSignal(false);
  const [communityBots, setCommunityBots] = createSignal<BotInfo[]>([]);
  const [communityBotsLoading, setCommunityBotsLoading] = createSignal(false);
  const [communitySearch, setCommunitySearch] = createSignal("");
  const [showCreateBot, setShowCreateBot] = createSignal(false);
  const [botName, setBotName] = createSignal("");
  const [botDesc, setBotDesc] = createSignal("");
  const [botWebhook, setBotWebhook] = createSignal("");
  const [showCredentials, setShowCredentials] = createSignal<{ bot_id: string; client_id: string; client_secret: string } | null>(null);
  const [editingBot, setEditingBot] = createSignal<BotInfo | null>(null);

  // Bot installation dialog state
  const [showInstallDialog, setShowInstallDialog] = createSignal(false);
  const [installBotId, setInstallBotId] = createSignal("");
  const [convList, setConvList] = createSignal<{ conv_id: string; name: string; type: string; avatar: string }[]>([]);
  const [selectedConvIds, setSelectedConvIds] = createSignal<Set<string>>(new Set());
  const [convListLoading, setConvListLoading] = createSignal(false);
  const [editBotName, setEditBotName] = createSignal("");
  const [editBotDesc, setEditBotDesc] = createSignal("");
  const [editBotWebhook, setEditBotWebhook] = createSignal("");

  const loadBots = async () => {
    setBotsLoading(true);
    try {
      const resp = await api.bot.listMyBots();
      setBots(resp.list || []);
    } catch {
      setBots([]);
    } finally {
      setBotsLoading(false);
    }
  };

  createResource(() => activeTab() === "bots", () => { if (activeTab() === "bots") loadBots(); });

  const loadCommunityBots = async (keyword?: string) => {
    setCommunityBotsLoading(true);
    try {
      const resp = await api.bot.listCommunityBots(keyword || undefined);
      setCommunityBots(resp.list || []);
    } catch {
      setCommunityBots([]);
    } finally {
      setCommunityBotsLoading(false);
    }
  };

  createResource(() => activeTab() === "community", () => { if (activeTab() === "community") loadCommunityBots(); });

  const handleInstallBot = (botId: string) => {
    setInstallBotId(botId);
    setSelectedConvIds(prev => new Set<string>());
    setShowInstallDialog(true);
    loadConvList();
  };

  const loadConvList = async () => {
    setConvListLoading(true);
    try {
      const resp = await api.conversation.list();
      setConvList(resp.list || []);
    } catch {
      setConvList([]);
    } finally {
      setConvListLoading(false);
    }
  };

  const toggleConvSelection = (convId: string) => {
    setSelectedConvIds(prev => {
      const next = new Set(prev);
      if (next.has(convId)) {
        next.delete(convId);
      } else {
        next.add(convId);
      }
      return next;
    });
  };

  const handleConfirmInstall = async () => {
    const convs = [...selectedConvIds()];
    if (convs.length === 0) {
      alert("请至少选择一个会话");
      return;
    }
    try {
      const resp = await api.bot.batchInstall(installBotId(), convs);
      if (resp.failed_convs && resp.failed_convs.length > 0) {
        alert(`安装完成：成功 ${resp.success_count} 个，失败 ${resp.failed_convs.length} 个`);
      } else {
        alert(`Bot 安装成功！已安装到 ${resp.success_count} 个会话`);
      }
      setShowInstallDialog(false);
    } catch {
      alert("Bot 安装失败，请稍后再试");
    }
  };

  const handleCreateBot = async (e: Event) => {
    e.preventDefault();
    if (!botName().trim()) return;
    try {
      const resp = await api.bot.createBot({
        name: botName().trim(),
        description: botDesc().trim(),
        webhook_url: botWebhook().trim(),
      } as CreateBotReq & { bot_id?: string });
      setShowCreateBot(false);
      setBotName("");
      setBotDesc("");
      setBotWebhook("");
      if (resp) {
        setShowCredentials({
          bot_id: (resp as unknown as Record<string,string>).bot_id || "",
          client_id: (resp as unknown as Record<string,string>).client_id || "",
          client_secret: (resp as unknown as Record<string,string>).client_secret || "",
        });
      }
      loadBots();
    } catch {
      // ignore
    }
  };

  const handleDeleteBot = async (botId: string) => {
    if (!confirm("确定要删除这个 Bot 吗？")) return;
    try {
      await api.bot.deleteBot(botId);
      loadBots();
    } catch {
      // ignore
    }
  };

  const handleEditBot = async (e: Event) => {
    e.preventDefault();
    const bot = editingBot();
    if (!bot) return;
    try {
      await api.bot.updateBot({
        bot_id: bot.bot_id,
        name: editBotName().trim() || bot.name,
        description: editBotDesc().trim() || bot.description,
        webhook_url: editBotWebhook().trim() || bot.webhook_url || "",
      });
      setEditingBot(null);
      loadBots();
    } catch {
      // ignore
    }
  };

  const startEditBot = (bot: BotInfo) => {
    setEditingBot(bot);
    setEditBotName(bot.name);
    setEditBotDesc(bot.description || "");
    setEditBotWebhook(bot.webhook_url || "");
  };

  const resolveUrl = (url?: string) => {
    if (!url) return "";
    if (url.startsWith("http://") || url.startsWith("https://")) return url;
    return getServerApiBase() + url;
  };

  return (
    <div class="h-screen flex flex-col bg-bg">
      {/* Header */}
      <div class="h-14 px-4 border-b border-border flex items-center gap-3 shrink-0 bg-bg-secondary/50">
        <button
          onClick={() => appNav.goChat()}
          class="p-1.5 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text"
        >
          <ArrowLeft size={18} />
        </button>
        <h1 class="text-base font-semibold text-text">设置</h1>
      </div>

      <div class="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div class="w-48 border-r border-border p-3 space-y-1 shrink-0">
          <button
            onClick={() => setActiveTab("general")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "general"
                ? "bg-primary/10 text-primary"
                : "text-text-secondary hover:bg-surface hover:text-text"
            }`}
          >
            <Palette size={16} />
            通用
          </button>
          <button
            onClick={() => setActiveTab("servers")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "servers"
                ? "bg-primary/10 text-primary"
                : "text-text-secondary hover:bg-surface hover:text-text"
            }`}
          >
            <Server size={16} />
            服务器
          </button>
          <button
            onClick={() => setActiveTab("account")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "account"
                ? "bg-primary/10 text-primary"
                : "text-text-secondary hover:bg-surface hover:text-text"
            }`}
          >
            <User size={16} />
            账号
          </button>
          <button
            onClick={() => setActiveTab("bots")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "bots"
                ? "bg-primary/10 text-primary"
                : "text-text-secondary hover:bg-surface hover:text-text"
            }`}
          >
            <Bot size={16} />
            Bot 管理
          </button>
          <button
            onClick={() => setActiveTab("community")}
            class={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeTab() === "community"
                ? "bg-primary/10 text-primary"
                : "text-text-secondary hover:bg-surface hover:text-text"
            }`}
          >
            <Globe size={16} />
            Bot 社区
          </button>
        </div>

        {/* Content */}
        <div class="flex-1 overflow-y-auto p-6">
          <Show when={activeTab() === "general"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">通用设置</h2>
                <div class="space-y-4">
                  <div class="bg-surface rounded-xl p-4 border border-border">
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                        <Palette size={18} class="text-text-muted" />
                        <div>
                          <p class="text-sm font-medium text-text">浅色模式</p>
                          <p class="text-xs text-text-muted">切换亮色/暗色主题</p>
                        </div>
                      </div>
                      <button
                        onClick={themeStore.toggleTheme}
                        class={`w-10 h-6 rounded-full relative cursor-pointer transition-colors ${
                          themeStore.theme() === "light" ? "bg-primary" : "bg-bg-tertiary"
                        }`}
                      >
                        <div
                          class={`w-4 h-4 bg-white rounded-full absolute top-1 transition-all ${
                            themeStore.theme() === "light" ? "right-1" : "left-1"
                          }`}
                        />
                      </button>
                    </div>
                  </div>

                  <div class="bg-surface rounded-xl p-4 border border-border">
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                        <Globe size={18} class="text-text-muted" />
                        <div>
                          <p class="text-sm font-medium text-text">语言</p>
                          <p class="text-xs text-text-muted">界面显示语言</p>
                        </div>
                      </div>
                      <select class="px-2 py-1 bg-bg border border-border rounded-lg text-xs text-text focus:outline-none focus:border-primary">
                        <option>简体中文</option>
                        <option>English</option>
                      </select>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Show>

          <Show when={activeTab() === "servers"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">服务器管理</h2>

                <div class="space-y-2 mb-4">
                  <For each={servers()}>
                    {(server) => (
                      <div
                        class={`flex items-center gap-3 px-4 py-3 bg-surface border rounded-xl transition-colors ${
                          activeServer()?.id === server.id
                            ? "border-primary/50 bg-primary/5"
                            : "border-border"
                        }`}
                      >
                        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                          <Server size={18} class="text-primary" />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium text-text">{server.name}</p>
                          <p class="text-xs text-text-muted truncate">{server.apiUrl}</p>
                        </div>
                        <Show when={activeServer()?.id === server.id}>
                          <span class="px-2 py-0.5 bg-primary/20 text-primary text-xs rounded-full font-medium">
                            当前
                          </span>
                        </Show>
                        <Show when={activeServer()?.id !== server.id}>
                          <button
                            onClick={() => handleSelectServer(server)}
                            class="px-3 py-1 text-xs text-primary hover:bg-primary/10 rounded-lg transition-colors"
                          >
                            切换
                          </button>
                        </Show>
                        <button
                          onClick={() => handleRemoveServer(server.id)}
                          class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-danger"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    )}
                  </For>
                </div>

                <form onSubmit={handleAddServer} class="flex gap-2">
                  <input
                    type="text"
                    value={newServerUrl()}
                    onInput={(e) => setNewServerUrl(e.currentTarget.value)}
                    placeholder="输入服务器地址..."
                    class="flex-1 px-3 py-2 bg-surface border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                  />
                  <button
                    type="submit"
                    disabled={addingServer() || !newServerUrl().trim()}
                    class="px-4 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl text-sm font-medium transition-colors shrink-0 flex items-center gap-1.5"
                  >
                    <Plus size={14} />
                    {addingServer() ? "添加中..." : "添加"}
                  </button>
                </form>
                <Show when={serverError()}>
                  <p class="text-xs text-danger mt-2">{serverError()}</p>
                </Show>
              </div>
            </div>
          </Show>

          <Show when={activeTab() === "account"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">账号信息</h2>
                <div class="bg-surface rounded-xl p-4 border border-border space-y-4">
                  <div class="flex items-center gap-4">
                    <label class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xl font-bold cursor-pointer hover:opacity-80 transition-opacity overflow-hidden relative group shrink-0">
                      <Show when={authStore.avatar()} fallback={<span>{authStore.name()?.charAt(0) || "U"}</span>}>
                        <img src={resolveUrl(authStore.avatar())} alt="" class="w-full h-full object-cover" />
                      </Show>
                      <span class="absolute inset-0 bg-black/40 items-center justify-center text-white text-xs hidden group-hover:flex rounded-full">
                        更换
                      </span>
                      <input
                        type="file"
                        accept="image/*"
                        class="hidden"
                        onChange={handleAvatarUpload}
                      />
                    </label>
                    <div class="flex-1">
                      <Show
                        when={editingName()}
                        fallback={
                          <div class="flex items-center gap-2">
                            <p class="text-base font-semibold text-text">{authStore.name() || "用户"}</p>
                            <button
                              onClick={() => {
                                setEditingName(true);
                                setNewName(authStore.name() || "");
                              }}
                              class="p-1 text-text-muted hover:text-text transition-colors"
                              title="修改昵称"
                            >
                              <Edit size={14} />
                            </button>
                          </div>
                        }
                      >
                        <div class="flex items-center gap-2">
                          <input
                            type="text"
                            value={newName()}
                            onInput={(e) => setNewName(e.currentTarget.value)}
                            class="flex-1 px-2 py-1 bg-bg border border-border rounded-lg text-sm text-text focus:outline-none focus:border-primary"
                            placeholder="输入新昵称"
                          />
                          <button
                            onClick={handleSaveName}
                            class="p-1.5 text-primary hover:bg-primary/10 rounded-lg transition-colors"
                            title="保存"
                          >
                            <Check size={14} />
                          </button>
                          <button
                            onClick={() => setEditingName(false)}
                            class="p-1.5 text-text-muted hover:text-text transition-colors"
                            title="取消"
                          >
                            <X size={14} />
                          </button>
                        </div>
                      </Show>
                      <p class="text-sm text-text-muted">UID: {authStore.uid()}</p>
                    </div>
                  </div>

                  <div class="border-t border-border pt-4 space-y-3">
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                        <Shield size={16} class="text-text-muted" />
                        <p class="text-sm text-text">安全设置</p>
                      </div>
                      <span class="text-xs text-text-muted">二次验证</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={handleLogout}
                  class="mt-4 w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-danger/10 hover:bg-danger/20 text-danger rounded-xl text-sm font-medium transition-colors"
                >
                  <LogOut size={16} />
                  退出登录
                </button>
              </div>
            </div>
          </Show>

          <Show when={activeTab() === "bots"}>
            <div class="max-w-lg space-y-6">
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-text">Bot 管理</h2>
                <button
                  onClick={() => setShowCreateBot(true)}
                  class="flex items-center gap-1.5 px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors"
                >
                  <Plus size={14} />
                  创建 Bot
                </button>
              </div>

              <Show when={botsLoading()}>
                <div class="text-center py-8 text-text-muted text-sm">加载中...</div>
              </Show>

              <Show when={!botsLoading() && bots().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm">
                  <Bot size={40} class="mx-auto mb-3 text-text-muted/30" />
                  <p>还没有创建任何 Bot</p>
                  <p class="text-xs mt-1">点击上方"创建 Bot"开始</p>
                </div>
              </Show>

              <div class="space-y-2">
                <For each={bots()}>
                  {(bot) => (
                    <div class="bg-surface rounded-xl p-4 border border-border">
                      <div class="flex items-center gap-3 mb-3">
                        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                          <Bot size={18} class="text-primary" />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-semibold text-text truncate">{bot.name}</p>
                          <p class="text-xs text-text-muted truncate">{bot.description || "暂无描述"}</p>
                        </div>
                        <div class="flex items-center gap-1">
                          <button
                            onClick={() => startEditBot(bot)}
                            class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text"
                            title="编辑"
                          >
                            <Edit size={14} />
                          </button>
                          <button
                            onClick={() => handleDeleteBot(bot.bot_id)}
                            class="p-1.5 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger"
                            title="删除"
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                      </div>
                      <div class="bg-bg rounded-lg p-2.5 space-y-1.5 text-xs font-mono">
                        <div class="flex items-center justify-between">
                          <span class="text-text-muted">Bot ID</span>
                          <span class="text-text">{bot.bot_id}</span>
                        </div>
                        <Show when={bot.webhook_url}>
                          <div class="flex items-center justify-between">
                            <span class="text-text-muted">Webhook</span>
                            <span class="text-text truncate max-w-[200px]">{bot.webhook_url}</span>
                          </div>
                        </Show>
                        <div class="flex items-center justify-between">
                          <span class="text-text-muted">创建时间</span>
                          <span class="text-text">{bot.created_at ? new Date(bot.created_at * 1000).toLocaleString() : "-"}</span>
                        </div>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>

          <Show when={activeTab() === "community"}>
            <div class="max-w-lg space-y-6">
              <div>
                <h2 class="text-lg font-semibold text-text mb-4">Bot 社区</h2>
                <p class="text-sm text-text-muted mb-4">发现和安装社区中优秀的 Bot</p>
                <div class="flex gap-2 mb-4">
                  <input
                    type="text"
                    value={communitySearch()}
                    onInput={(e) => setCommunitySearch(e.currentTarget.value)}
                    placeholder="搜索 Bot..."
                    class="flex-1 px-3 py-2 bg-surface border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                    onKeyDown={(e) => { if (e.key === "Enter") loadCommunityBots(communitySearch()); }}
                  />
                  <button
                    onClick={() => loadCommunityBots(communitySearch())}
                    class="px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors shrink-0"
                  >
                    搜索
                  </button>
                </div>
              </div>

              <Show when={communityBotsLoading()}>
                <div class="text-center py-8 text-text-muted text-sm">加载中...</div>
              </Show>

              <Show when={!communityBotsLoading() && communityBots().length === 0}>
                <div class="text-center py-12 text-text-muted text-sm">
                  <Globe size={40} class="mx-auto mb-3 text-text-muted/30" />
                  <p>暂无社区 Bot</p>
                  <p class="text-xs mt-1">尝试搜索其他关键词</p>
                </div>
              </Show>

              <div class="space-y-2">
                <For each={communityBots()}>
                  {(bot) => (
                    <div class="bg-surface rounded-xl p-4 border border-border">
                      <div class="flex items-center gap-3 mb-2">
                        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                          <Bot size={18} class="text-primary" />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-semibold text-text truncate">{bot.name}</p>
                          <p class="text-xs text-text-muted truncate">{bot.description || "暂无描述"}</p>
                        </div>
                        <button
                          onClick={() => handleInstallBot(bot.bot_id)}
                          class="px-3 py-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors shrink-0 flex items-center gap-1"
                        >
                          <Plus size={12} />
                          安装
                        </button>
                      </div>
                      <div class="bg-bg rounded-lg p-2.5 text-xs text-text-muted">
                        Bot ID: {bot.bot_id}
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>
        </div>
      </div>

      {/* Create Bot Modal */}
      <Show when={showCreateBot()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowCreateBot(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">创建 Bot</h2>
            <form onSubmit={handleCreateBot} class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">名称</label>
                <input
                  type="text"
                  value={botName()}
                  onInput={(e) => setBotName(e.currentTarget.value)}
                  placeholder="Bot 名称"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                  required
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">描述</label>
                <input
                  type="text"
                  value={botDesc()}
                  onInput={(e) => setBotDesc(e.currentTarget.value)}
                  placeholder="简要描述 Bot 功能"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">Webhook URL (可选)</label>
                <input
                  type="url"
                  value={botWebhook()}
                  onInput={(e) => setBotWebhook(e.currentTarget.value)}
                  placeholder="https://your-bot.example.com/webhook"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div class="flex gap-2 pt-1">
                <button type="button" onClick={() => setShowCreateBot(false)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">创建</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Credentials Modal */}
      <Show when={showCredentials()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowCredentials(null)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-2">Bot 创建成功</h2>
            <p class="text-xs text-text-muted mb-4">请保存以下凭证，Client Secret 仅显示一次。</p>
            <div class="bg-bg rounded-xl p-3 space-y-2 text-xs font-mono mb-4">
              <div>
                <span class="text-text-muted">Bot ID: </span>
                <span class="text-text">{showCredentials()?.bot_id}</span>
              </div>
              <div>
                <span class="text-text-muted">Client ID: </span>
                <span class="text-text break-all">{showCredentials()?.client_id}</span>
              </div>
              <div>
                <span class="text-text-muted">Client Secret: </span>
                <span class="text-warning break-all">{showCredentials()?.client_secret}</span>
              </div>
            </div>
            <button
              onClick={() => setShowCredentials(null)}
              class="w-full px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors"
            >
              我已保存
            </button>
          </div>
        </div>
      </Show>

      {/* Edit Bot Modal */}
      <Show when={editingBot()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setEditingBot(null)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">编辑 Bot</h2>
            <form onSubmit={handleEditBot} class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">名称</label>
                <input
                  type="text"
                  value={editBotName()}
                  onInput={(e) => setEditBotName(e.currentTarget.value)}
                  placeholder="Bot 名称"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">描述</label>
                <input
                  type="text"
                  value={editBotDesc()}
                  onInput={(e) => setEditBotDesc(e.currentTarget.value)}
                  placeholder="简要描述 Bot 功能"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-text-muted mb-1">Webhook URL (可选)</label>
                <input
                  type="url"
                  value={editBotWebhook()}
                  onInput={(e) => setEditBotWebhook(e.currentTarget.value)}
                  placeholder="https://your-bot.example.com/webhook"
                  class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div class="flex gap-2 pt-1">
                <button type="button" onClick={() => setEditingBot(null)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">保存</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Install Bot Dialog */}
      <Show when={showInstallDialog()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowInstallDialog(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[80vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-2">选择安装会话</h2>
            <p class="text-xs text-text-muted mb-3">请选择要安装 Bot 的会话（可多选）</p>
            <Show when={convListLoading()}>
              <p class="text-sm text-text-muted text-center py-8">加载中...</p>
            </Show>
            <Show when={!convListLoading()}>
              <div class="flex-1 overflow-y-auto space-y-1 mb-4">
                <For each={convList()}>
                  {(conv) => (
                    <div
                      class={`flex items-center gap-3 p-3 rounded-xl cursor-pointer transition-colors ${selectedConvIds().has(conv.conv_id) ? "bg-primary/10 border border-primary/30" : "hover:bg-bg border border-transparent"}`}
                      onClick={() => toggleConvSelection(conv.conv_id)}
                    >
                      <div class={`w-5 h-5 rounded-md border-2 flex items-center justify-center shrink-0 transition-colors ${selectedConvIds().has(conv.conv_id) ? "bg-primary border-primary" : "border-border"}`}>
                        <Show when={selectedConvIds().has(conv.conv_id)}>
                          <Check size={12} class="text-white" />
                        </Show>
                      </div>
                      <div class="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                        <span class="text-xs text-primary font-bold">{conv.name.charAt(0)}</span>
                      </div>
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-text truncate">{conv.name}</p>
                        <p class="text-xs text-text-muted">{conv.type === "group" ? "群聊" : "单聊"}</p>
                      </div>
                    </div>
                  )}
                </For>
                <Show when={convList().length === 0}>
                  <p class="text-sm text-text-muted text-center py-8">暂无会话</p>
                </Show>
              </div>
            </Show>
            <div class="flex gap-2 pt-2 border-t border-border">
              <button onClick={() => setShowInstallDialog(false)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
              <button onClick={handleConfirmInstall} class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors" disabled={convListLoading()}>
                安装到 {selectedConvIds().size} 个会话
              </button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  );
}