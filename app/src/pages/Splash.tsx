import { createSignal, onMount, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { authStore } from "../stores/auth";
import { api, validateToken, clearTokens, checkServerHealth } from "../services/api";
import { getActiveServer, getSavedServers, addServer, setActiveServerId, generateServerId, parseServerUrl } from "../services/config";
import { MessageSquare } from "lucide-solid";

export function SplashScreen() {
  const navigate = useNavigate();
  const [status, setStatus] = createSignal("正在检查服务器...");
  const [error, setError] = createSignal("");
  const [showServerInput, setShowServerInput] = createSignal(false);
  const [serverUrl, setServerUrl] = createSignal("");
  const [loading, setLoading] = createSignal(false);

  onMount(async () => {
    await checkServerAndLogin();
  });

  const checkServerAndLogin = async () => {
    setError("");
    
    const savedServers = getSavedServers();
    if (savedServers.length === 0) {
      setShowServerInput(true);
      setStatus("请添加服务器地址");
      return;
    }

    const activeServer = getActiveServer();
    if (!activeServer) {
      setShowServerInput(true);
      setStatus("请选择服务器地址");
      return;
    }

    setStatus("正在连接服务器...");
    try {
      const isHealthy = await checkServerHealth(activeServer.apiUrl);
      if (!isHealthy) {
        setError(`无法连接到服务器: ${activeServer.apiUrl}`);
        setShowServerInput(true);
        return;
      }
    } catch (err) {
      setError(`服务器连接失败: ${err instanceof Error ? err.message : "未知错误"}`);
      setShowServerInput(true);
      return;
    }

    setStatus("正在验证登录状态...");
    const isTokenValid = await validateToken();
    
    if (isTokenValid && authStore.isLoggedIn()) {
      setStatus("正在加载用户信息...");
      await authStore.fetchUserInfo();
      setStatus("登录成功，正在跳转...");
      setTimeout(() => {
        navigate("/chat", { replace: true });
      }, 500);
    } else {
      clearTokens();
      setStatus("需要登录");
      setTimeout(() => {
        navigate("/login", { replace: true });
      }, 800);
    }
  };

  const handleAddServer = async (e: Event) => {
    e.preventDefault();
    if (!serverUrl().trim()) return;

    setLoading(true);
    setError("");

    try {
      const { apiUrl, wsUrl, name } = parseServerUrl(serverUrl());
      
      // 测试服务器连接
      setStatus("正在验证服务器...");
      const isHealthy = await checkServerHealth(apiUrl);
      if (!isHealthy) {
        setError("无法连接到该服务器，请检查地址是否正确");
        setLoading(false);
        return;
      }

      // 添加服务器
      const server = {
        id: generateServerId(),
        name,
        apiUrl,
        wsUrl,
      };
      addServer(server);
      setActiveServerId(server.id);

      setStatus("服务器添加成功，正在重试...");
      setTimeout(() => {
        checkServerAndLogin();
      }, 1000);
    } catch (err) {
      setError(`服务器配置失败: ${err instanceof Error ? err.message : "未知错误"}`);
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
        </div>

        <Show
          when={!showServerInput()}
          fallback={
            <div>
              <form onSubmit={handleAddServer} class="space-y-4">
                <div class="bg-surface rounded-2xl p-6 border border-border">
                  <h2 class="text-lg font-semibold text-text mb-1">添加服务器</h2>
                  <p class="text-sm text-text-muted mb-4">
                    请输入服务器地址（如: localhost:80 或 192.168.1.100:8080）
                  </p>
                  <input
                    type="text"
                    value={serverUrl()}
                    onInput={(e) => setServerUrl(e.currentTarget.value)}
                    placeholder="服务器地址"
                    class="w-full px-3 py-2 bg-bg border border-border rounded-lg text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                    required
                  />
                </div>
                
                <Show when={error()}>
                  <p class="text-sm text-danger text-center">{error()}</p>
                </Show>
                
                <button
                  type="submit"
                  disabled={loading()}
                  class="w-full px-4 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl transition-all font-medium"
                >
                  {loading() ? "连接中..." : "连接"}
                </button>
              </form>
              
              <div class="mt-4 text-center">
                <p class="text-xs text-text-muted">
                  服务器地址将保存在本地，不会上传到任何地方
                </p>
              </div>
            </div>
          }
        >
          <div class="text-center space-y-4">
            <div class="bg-surface rounded-2xl p-6 border border-border">
              <div class="flex items-center justify-center gap-3 mb-3">
                <div class="w-3 h-3 bg-primary rounded-full animate-pulse" />
                <p class="text-sm text-text">{status()}</p>
              </div>
              
              <Show when={error()}>
                <p class="text-sm text-danger mt-2">{error()}</p>
                <div class="flex gap-2 mt-3 justify-center">
                  <button
                    onClick={() => setShowServerInput(true)}
                    class="text-sm text-primary hover:text-primary-dark transition-colors"
                  >
                    更换服务器
                  </button>
                  <span class="text-text-muted">|</span>
                  <button
                    onClick={() => navigate("/login", { replace: true })}
                    class="text-sm text-primary hover:text-primary-dark transition-colors"
                  >
                    直接登录
                  </button>
                </div>
              </Show>
            </div>
            
            <button
              onClick={() => setShowServerInput(true)}
              class="text-sm text-text-muted hover:text-text transition-colors"
            >
              管理服务器
            </button>
          </div>
        </Show>
      </div>
    </div>
  );
}