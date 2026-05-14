import { createSignal, Show, onMount, For } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { authStore } from "../stores/auth";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { MessageSquare, Server, ChevronDown, X, Check } from "lucide-solid";
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

export function RegisterPage() {
  const navigate = useNavigate();
  const [phone, setPhone] = createSignal("");
  const [name, setName] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal("");
  const [servers, setServers] = createSignal<ServerConfig[]>([]);
  const [activeServer, setActiveServer] = createSignal<ServerConfig | null>(null);
  const [showServerManager, setShowServerManager] = createSignal(false);
  const [newServerUrl, setNewServerUrl] = createSignal("");
  const [addingServer, setAddingServer] = createSignal(false);
  const [serverError, setServerError] = createSignal("");

  onMount(() => {
    const saved = getSavedServers();
    setServers(saved);
    const active = getActiveServer();
    setActiveServer(active);

    if (saved.length === 0) {
      setShowServerManager(true);
    }
  });

  const handleSelectServer = (server: ServerConfig) => {
    setActiveServerId(server.id);
    setActiveServer(server);
    setShowServerManager(false);
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
      setShowServerManager(false);
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

  const handleRegister = async (e: Event) => {
    e.preventDefault();
    setError("");

    if (!activeServer()) {
      setError("请先选择服务器");
      return;
    }

    if (!phone().trim() || !name().trim() || !password().trim()) {
      setError("请填写所有字段");
      return;
    }

    setLoading(true);

    try {
      await authStore.register(phone(), name(), password());
      navigate("/login", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "注册失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="min-h-screen flex items-center justify-center bg-bg p-4">
      <div class="w-full max-w-sm">
        <div class="text-center mb-8">
          <div class="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto mb-4">
            <MessageSquare size={32} class="text-primary" />
          </div>
          <h1 class="text-2xl font-bold text-text">Katheryne</h1>
          <p class="text-text-muted text-sm mt-1">创建新账号</p>
        </div>

        {/* Server Selector */}
        <div class="mb-4">
          <button
            onClick={() => setShowServerManager(!showServerManager())}
            class="w-full flex items-center gap-2 px-4 py-2.5 bg-surface border border-border rounded-xl text-sm text-text hover:border-primary/50 transition-colors"
          >
            <Server size={16} class="text-text-muted" />
            <span class="flex-1 text-left truncate">
              {activeServer() ? activeServer()!.name : "选择服务器"}
            </span>
            <ChevronDown
              size={16}
              class={`text-text-muted transition-transform ${showServerManager() ? "rotate-180" : ""}`}
            />
          </button>

          <Show when={showServerManager()}>
            <div class="mt-2 bg-surface border border-border rounded-xl overflow-hidden">
              <div class="max-h-48 overflow-y-auto">
                <Show when={servers().length > 0}>
                  <For each={servers()}>
                    {(server) => (
                      <div
                        onClick={() => handleSelectServer(server)}
                        class={`w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors hover:bg-bg cursor-pointer ${
                          activeServer()?.id === server.id ? "bg-primary/5" : ""
                        }`}
                      >
                        <div class="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                          <Server size={14} class="text-primary" />
                        </div>
                        <div class="flex-1 text-left min-w-0">
                          <p class="text-text font-medium truncate">{server.name}</p>
                          <p class="text-xs text-text-muted truncate">{server.apiUrl}</p>
                        </div>
                        <Show when={activeServer()?.id === server.id}>
                          <Check size={16} class="text-primary shrink-0" />
                        </Show>
                        <span
                          onClick={(e) => {
                            e.stopPropagation();
                            handleRemoveServer(server.id);
                          }}
                          class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted hover:text-danger shrink-0 cursor-pointer"
                          role="button"
                        >
                          <X size={14} />
                        </span>
                      </div>
                    )}
                  </For>
                </Show>
                <Show when={servers().length === 0}>
                  <div class="px-4 py-6 text-center text-sm text-text-muted">
                    暂无保存的服务器
                  </div>
                </Show>
              </div>

              <div class="border-t border-border p-3">
                <form onSubmit={handleAddServer} class="flex gap-2">
                  <input
                    type="text"
                    value={newServerUrl()}
                    onInput={(e) => setNewServerUrl(e.currentTarget.value)}
                    placeholder="输入服务器地址..."
                    class="flex-1 px-3 py-1.5 bg-bg border border-border rounded-lg text-xs text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                  />
                  <button
                    type="submit"
                    disabled={addingServer() || !newServerUrl().trim()}
                    class="px-3 py-1.5 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg text-xs font-medium transition-colors shrink-0"
                  >
                    {addingServer() ? "..." : "添加"}
                  </button>
                </form>
                <Show when={serverError()}>
                  <p class="text-xs text-danger mt-2">{serverError()}</p>
                </Show>
              </div>
            </div>
          </Show>
        </div>

        <form onSubmit={handleRegister} class="space-y-4">
          <div class="bg-surface rounded-2xl p-6 border border-border space-y-4">
            <Input
              label="手机号"
              type="tel"
              value={phone()}
              onInput={(e) => setPhone(e.currentTarget.value)}
              placeholder="输入手机号"
              required
            />
            <Input
              label="昵称"
              type="text"
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              placeholder="输入昵称"
              required
            />
            <Input
              label="密码"
              type="password"
              value={password()}
              onInput={(e) => setPassword(e.currentTarget.value)}
              placeholder="输入密码"
              required
            />
          </div>
          <Show when={error()}>
            <p class="text-sm text-danger text-center">{error()}</p>
          </Show>
          <Button type="submit" class="w-full" disabled={loading() || !activeServer()}>
            {loading() ? "注册中..." : "注册"}
          </Button>
        </form>

        <p class="text-center text-sm text-text-muted mt-6">
          已有账号？{" "}
          <a
            href="/login"
            class="text-primary hover:text-primary-light transition-colors font-medium"
            onClick={(e) => {
              e.preventDefault();
              navigate("/login");
            }}
          >
            登录
          </a>
        </p>
      </div>
    </div>
  );
}