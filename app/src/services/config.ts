const STORAGE_KEY = "katheryne_servers";
const ACTIVE_SERVER_KEY = "katheryne_active_server";

export interface ServerConfig {
  id: string;
  name: string;
  apiUrl: string;
  wsUrl: string;
}

export function getSavedServers(): ServerConfig[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function saveServers(servers: ServerConfig[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(servers));
}

export function addServer(server: ServerConfig) {
  const servers = getSavedServers();
  const existing = servers.findIndex((s) => s.apiUrl === server.apiUrl);
  if (existing >= 0) {
    servers[existing] = server;
  } else {
    servers.push(server);
  }
  saveServers(servers);
}

export function removeServer(id: string) {
  const servers = getSavedServers().filter((s) => s.id !== id);
  saveServers(servers);
  if (getActiveServerId() === id) {
    clearActiveServer();
  }
}

export function getActiveServerId(): string | null {
  return localStorage.getItem(ACTIVE_SERVER_KEY);
}

export function setActiveServerId(id: string) {
  localStorage.setItem(ACTIVE_SERVER_KEY, id);
}

export function clearActiveServer() {
  localStorage.removeItem(ACTIVE_SERVER_KEY);
}

export function getActiveServer(): ServerConfig | null {
  const id = getActiveServerId();
  if (!id) return null;
  return getSavedServers().find((s) => s.id === id) || null;
}

export function getServerApiBase(): string {
  const server = getActiveServer();
  return server?.apiUrl || "http://localhost:80";
}

export function getServerWsUrl(): string {
  const server = getActiveServer();
  let url = server?.wsUrl || "ws://localhost:8080/api/v1/ws";
  if (url.endsWith("/ws") && !url.includes("/api/v1/ws")) {
    url = url.replace(/\/ws$/, "/api/v1/ws");
  }
  return url;
}

export function generateServerId(): string {
  return `srv_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

export function parseServerUrl(input: string): { apiUrl: string; wsUrl: string; name: string } {
  let url = input.trim();
  if (!url.startsWith("http://") && !url.startsWith("https://")) {
    url = "http://" + url;
  }

  const parsed = new URL(url);
  const host = parsed.hostname;
  const port = parsed.port;
  const protocol = parsed.protocol;

  const apiPort = port || "80";
  const wsPort = port || "8080";
  const wsProtocol = protocol === "https:" ? "wss:" : "ws:";

  return {
    apiUrl: `${protocol}//${host}:${apiPort}`,
    wsUrl: `${wsProtocol}//${host}:${wsPort}/api/v1/ws`,
    name: host,
  };
}