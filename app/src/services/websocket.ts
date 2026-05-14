import { getAccessToken, refreshAccessToken, triggerAuthFailed } from "./api";
import { getServerWsUrl } from "./config";

type MessageHandler = (data: WSPushMessage) => void;

export interface WSPushMessage {
  type: string;
  data: unknown;
  seq?: number;
}

export interface NewMessagePush {
  msg_id: string;
  conv_id: string;
  sender: string;
  sender_name?: string;
  sender_avatar?: string;
  receiver: string;
  msg_type: string;
  content: string;
  content_type: string;
  quote_msg_id?: string;
  extra?: string;
  created_at: number;
}

export interface ReadReceiptPush {
  conv_id: string;
  uid: string;
  last_read_msg_id: string;
}

export interface TypingPush {
  conv_id: string;
  uid: string;
  status: string;
}

export interface RecallMessagePush {
  conv_id: string;
  msg_id: string;
  uid: string;
}

export interface EditMessagePush {
  conv_id: string;
  msg_id: string;
  content: string;
  uid: string;
}

export interface OnlineStatusPush {
  uid: string;
  status: string;
}

let instanceCounter = 0;

class WebSocketService {
  private _instanceId: number;
  private ws: WebSocket | null = null;
  private _wsId = 0;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private seq = 0;
  private authenticated = false;
  private uid = "";
  private shouldReconnect = true;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;
  private _connecting = false;

  constructor() {
    this._instanceId = ++instanceCounter;
    console.log("[WS] WebSocketService instance created, id:", this._instanceId);
  }

  private _captureStack(): string {
    const stack = new Error().stack || "";
    const lines = stack.split("\n").slice(2, 6);
    return lines.map((l) => l.trim()).join(" <- ");
  }

  connect() {
    const stack = this._captureStack();
    const token = getAccessToken();
    console.log("[WS] connect() called, instance:", this._instanceId, "token exists:", !!token, "ws readyState:", this.ws?.readyState, "_connecting:", this._connecting, "caller:", stack);

    if (this.ws?.readyState === WebSocket.OPEN) {
      console.warn("[WS] connect() aborted: already OPEN, instance:", this._instanceId, "caller:", stack);
      return;
    }

    if (!token) {
      console.warn("[WS] connect() aborted: no token, instance:", this._instanceId);
      return;
    }

    if (this._connecting) {
      console.warn("[WS] connect() aborted: already connecting, instance:", this._instanceId, "caller:", stack);
      return;
    }

    if (this.ws?.readyState === WebSocket.CONNECTING) {
      console.warn("[WS] connect() aborted: already CONNECTING, instance:", this._instanceId, "caller:", stack);
      return;
    }

    if (this.ws) {
      console.warn("[WS] connect() cleaning up existing ws in state:", this.ws.readyState, "instance:", this._instanceId);
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.close();
      this.ws = null;
    }

    this._connecting = true;
    this.shouldReconnect = true;
    this._wsId++;
    const wsId = this._wsId;
    const wsUrl = getServerWsUrl();
    console.log("[WS] connect() creating ws#", wsId, "instance:", this._instanceId, "url:", wsUrl);
    this.ws = new WebSocket(`${wsUrl}?token=${encodeURIComponent(token)}`);
    const currentWs = this.ws;

    this.ws.onopen = () => {
      console.log("[WS] onopen fired, ws#", wsId, "instance:", this._instanceId, "readyState:", this.ws?.readyState);
      this._connecting = false;
      this.authenticated = true;
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: WSPushMessage = JSON.parse(event.data);
        if (msg.type === "auth_resp") {
          const data = msg.data as { success: boolean; uid?: string };
          console.log("[WS] auth_resp received, ws#", wsId, "instance:", this._instanceId, "data:", data);
          if (data.success) {
            this.uid = data.uid || "";
          } else {
            this.authenticated = false;
          }
          return;
        }
        this.dispatch(msg);
      } catch {
        // ignore parse errors
      }
    };

    this.ws.onclose = (event) => {
      if (this.ws !== currentWs) {
        console.log("[WS] onclose ignored: stale ws#", wsId, "instance:", this._instanceId, "current readyState:", this.ws?.readyState);
        return;
      }
      this._connecting = false;
      console.log("[WS] onclose fired, ws#", wsId, "instance:", this._instanceId, "code:", event.code, "reason:", event.reason, "wasAuthenticated:", this.authenticated, "shouldReconnect:", this.shouldReconnect);
      const wasAuthenticated = this.authenticated;
      this.authenticated = false;
      if (this.shouldReconnect) {
        if (!wasAuthenticated && this.reconnectAttempts === 0) {
          console.log("[WS] onclose -> tryRefreshAndReconnect, ws#", wsId, "instance:", this._instanceId);
          this.tryRefreshAndReconnect();
        } else {
          console.log("[WS] onclose -> scheduleReconnect, ws#", wsId, "instance:", this._instanceId);
          this.scheduleReconnect();
        }
      }
    };

    this.ws.onerror = (event) => {
      console.error("[WS] onerror fired, ws#", wsId, "instance:", this._instanceId);
      this.ws?.close();
    };
  }

  private async tryRefreshAndReconnect() {
    this.reconnectAttempts++;
    if (this.reconnectAttempts > this.maxReconnectAttempts) {
      triggerAuthFailed();
      return;
    }
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      this.reconnectAttempts = 0;
      this.connect();
    } else {
      this.shouldReconnect = false;
      triggerAuthFailed();
    }
  }

  disconnect() {
    const stack = this._captureStack();
    console.log("[WS] disconnect() called, instance:", this._instanceId, "ws readyState:", this.ws?.readyState, "_connecting:", this._connecting, "caller:", stack);
    this.shouldReconnect = false;
    this._connecting = false;
    this.clearReconnect();
    this.authenticated = false;
    if (this.ws) {
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.close();
      this.ws = null;
    }
  }

  private scheduleReconnect() {
    this.clearReconnect();
    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, 3000);
  }

  private clearReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  send(data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const msg = { ...data, seq: ++this.seq };
      console.log("[WS] send() type:", data.type, "seq:", msg.seq, "instance:", this._instanceId);
      this.ws.send(JSON.stringify(msg));
    } else {
      console.warn("[WS] send() skipped: ws not OPEN, readyState:", this.ws?.readyState, "authenticated:", this.authenticated, "type:", data.type, "instance:", this._instanceId);
    }
  }

  sendMessage(data: {
    conv_id: string;
    receiver: string;
    msg_type: string;
    content: string;
    content_type: string;
    quote_msg_id?: string;
    extra?: string;
    temp_id?: string;
  }) {
    this.send({ type: "send_message", data });
  }

  sendReadReceipt(convId: string, lastReadMsgId: string, startMsgId?: string, endMsgId?: string) {
    this.send({
      type: "read_receipt",
      data: { conv_id: convId, last_read_msg_id: lastReadMsgId, start_msg_id: startMsgId, end_msg_id: endMsgId },
    });
  }

  sendTyping(convId: string, receiver: string, status: string) {
    this.send({
      type: "typing",
      data: { conv_id: convId, receiver, status },
    });
  }

  sendRecallMessage(convId: string, msgId: string) {
    this.send({
      type: "recall_message",
      data: { conv_id: convId, msg_id: msgId },
    });
  }

  sendEditMessage(convId: string, msgId: string, content: string) {
    this.send({
      type: "edit_message",
      data: { conv_id: convId, msg_id: msgId, content },
    });
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  private dispatch(msg: WSPushMessage) {
    const handlers = this.handlers.get(msg.type);
    console.log("[WS] dispatch() type:", msg.type, "handlerCount:", handlers?.size || 0, "instance:", this._instanceId);
    if (handlers) {
      handlers.forEach((h) => h(msg));
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN && this.authenticated;
  }

  getUid(): string {
    return this.uid;
  }
}

export const wsService = new WebSocketService();