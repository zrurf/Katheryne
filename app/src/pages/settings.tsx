import { createSignal, Show, onMount, For } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { authStore } from "../stores/auth";
import {
  getSavedServers,
  getActiveServer,
  setActiveServerId,
  addServer,
  removeServer,
  generateServerId,
  parseServerUrl,
  type ServerConfig,
} from "../services/config";
import { checkServerHealth } from "../services/api";
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
} from "lucide-solid";

export function SettingsPage() {
  const navigate = useNavigate();
  const [servers, setServers] = createSignal<ServerConfig[]>([]);
  const [activeServer, setActiveServer] = createSignal<ServerConfig | null>(null);
  const [newServerUrl, setNewServerUrl] = createSignal("");
  const [addingServer, setAddingServer] = createSignal(false);
  const [serverError, setServerError] = createSignal("");
  const [activeTab, setActiveTab] = createSignal<"general" | "servers" | "account">("general");

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

  return (
    <div class="h-screen flex flex-col bg-bg">
      {/* Header */}
      <div class="h-14 px-4 border-b border-border flex items-center gap-3 shrink-0 bg-bg-secondary/50">
        <button
          onClick={() => navigate("/chat")}
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
                        <Bell size={18} class="text-text-muted" />
                        <div>
                          <p class="text-sm font-medium text-text">消息通知</p>
                          <p class="text-xs text-text-muted">接收新消息通知</p>
                        </div>
                      </div>
                      <div class="w-10 h-6 bg-primary rounded-full relative cursor-pointer">
                        <div class="w-4 h-4 bg-white rounded-full absolute top-1 right-1 transition-all" />
                      </div>
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
                    <div class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xl font-bold">
                      {authStore.name()?.charAt(0) || "U"}
                    </div>
                    <div>
                      <p class="text-base font-semibold text-text">{authStore.name() || "用户"}</p>
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
        </div>
      </div>
    </div>
  );
}