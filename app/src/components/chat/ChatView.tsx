import { For, Show, createSignal, onCleanup, onMount, createEffect } from "solid-js";
import { chatStore } from "../../stores/chat";
import { authStore } from "../../stores/auth";
import { Avatar } from "../ui/avatar";
import { GroupInfoPanel } from "./GroupInfoPanel";
import { formatTime } from "../../lib/utils";
import type { MessageItem } from "../../services/api";
import {
  Smile,
  Paperclip,
  Image,
  Send,
  Reply,
  X,
  MoreVertical,
  Phone,
  Video,
  Info,
  Check,
  CheckCheck,
  Edit3,
  ArrowUp,
  Eye,
} from "lucide-solid";

export function ChatView() {
  const [inputText, setInputText] = createSignal("");
  const [replyTo, setReplyTo] = createSignal<MessageItem | null>(null);
  const [editingMsg, setEditingMsg] = createSignal<MessageItem | null>(null);
  let messagesEndRef: HTMLDivElement | undefined;
  let messagesContainerRef: HTMLDivElement | undefined;
  let inputRef: HTMLTextAreaElement | undefined;
  let typingTimer: ReturnType<typeof setTimeout> | null = null;
  let isTyping = false;
  let readTimer: ReturnType<typeof setTimeout> | null = null;
  let lastSubmittedEndMsgId = "";

  const activeConv = () => {
    const id = chatStore.activeConvId();
    return chatStore.conversations().find((c) => c.conv_id === id);
  };

  const scrollToBottom = () => {
    messagesEndRef?.scrollIntoView({ behavior: "smooth" });
  };

  onMount(() => {
    scrollToBottom();
  });

  onCleanup(() => {
    if (typingTimer) clearTimeout(typingTimer);
    if (readTimer) clearTimeout(readTimer);
  });

  createEffect(() => {
    const msgs = chatStore.messages();
    if (msgs.length > 0) {
      scrollToBottom();
      setTimeout(() => trackVisibleMessages(), 100);
    }
  });

  createEffect(() => {
    chatStore.activeConvId();
    lastSubmittedEndMsgId = "";
  });

  const handleScroll = () => {
    const container = messagesContainerRef;
    if (!container) return;
    if (container.scrollTop < 50 && chatStore.hasMore() && !chatStore.loading()) {
      chatStore.loadMoreMessages();
    }
    trackVisibleMessages();
  };

  const trackVisibleMessages = () => {
    const container = messagesContainerRef;
    if (!container) return;

    const conv = activeConv();
    if (!conv) return;

    const msgElements = container.querySelectorAll("[data-msg-id]");
    if (msgElements.length === 0) return;

    const containerRect = container.getBoundingClientRect();
    let minMsgId = "";
    let maxMsgId = "";

    msgElements.forEach((el) => {
      const rect = el.getBoundingClientRect();
      if (rect.bottom > containerRect.top && rect.top < containerRect.bottom) {
        const msgId = el.getAttribute("data-msg-id");
        if (msgId) {
          if (!minMsgId || msgId < minMsgId) minMsgId = msgId;
          if (!maxMsgId || msgId > maxMsgId) maxMsgId = msgId;
        }
      }
    });

    if (!minMsgId || !maxMsgId) return;

    if (readTimer) clearTimeout(readTimer);
    readTimer = setTimeout(() => {
      if (maxMsgId === lastSubmittedEndMsgId) return;
      lastSubmittedEndMsgId = maxMsgId;
      chatStore.sendReadReceipt(conv.conv_id, maxMsgId, minMsgId, maxMsgId);
    }, 500);
  };

  const handleSend = () => {
    const text = inputText().trim();
    if (!text) return;

    const conv = activeConv();
    if (!conv) return;

    if (editingMsg()) {
      chatStore.editMessage(conv.conv_id, editingMsg()!.id, text);
      setEditingMsg(null);
    } else {
      chatStore.sendMessage(
        conv.conv_id,
        "",
        "text",
        text,
        "text",
        replyTo()?.id
      );
    }
    setInputText("");
    setReplyTo(null);
    scrollToBottom();
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleRecall = (msg: MessageItem) => {
    chatStore.recallMessage(msg.conv_id, msg.id);
  };

  const handleEdit = (msg: MessageItem) => {
    setEditingMsg(msg);
    setInputText(msg.content);
    setReplyTo(null);
    inputRef?.focus();
  };

  const handleInputTyping = () => {
    const conv = activeConv();
    if (!conv) return;

    if (!isTyping) {
      isTyping = true;
      chatStore.sendTyping(conv.conv_id, "", "typing");
    }

    if (typingTimer) clearTimeout(typingTimer);
    typingTimer = setTimeout(() => {
      isTyping = false;
      chatStore.sendTyping(conv.conv_id, "", "stop_typing");
    }, 3000);
  };

  const typingIndicator = () => {
    const users = chatStore.typingUsers();
    const currentUid = authStore.uid();
    const typingUids = Object.keys(users).filter((uid) => uid !== currentUid && users[uid] === "typing");
    if (typingUids.length === 0) return null;
    return "对方正在输入...";
  };

  return (
    <div class="flex-1 flex h-full bg-bg">
      <div class="flex-1 flex flex-col min-w-0">
        <Show
          when={activeConv()}
          fallback={
            <div class="flex-1 flex items-center justify-center">
              <div class="text-center">
                <div class="w-20 h-20 rounded-full bg-surface flex items-center justify-center mx-auto mb-4">
                  <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-primary">
                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                  </svg>
                </div>
                <h2 class="text-xl font-semibold text-text mb-2">
                  欢迎使用 Katheryne
                </h2>
                <p class="text-text-muted text-sm">
                  选择一个会话开始聊天
                </p>
              </div>
            </div>
          }
        >
          {/* Chat Header */}
          <div class="h-16 px-5 border-b border-border flex items-center justify-between shrink-0 bg-bg-secondary/50 backdrop-blur-sm">
            <div class="flex items-center gap-3">
              <Avatar
                name={activeConv()?.name}
                src={activeConv()?.avatar}
                size="md"
              />
              <div>
                <h3 class="text-sm font-semibold text-text">
                  {activeConv()?.name}
                </h3>
                <p class="text-xs text-text-muted">
                  {typingIndicator() || (activeConv()?.type === "GROUP" ? `${chatStore.groupInfo()?.member_count || 0} 名成员` : (() => {
                    const peerUid = activeConv()?.peer_uid;
                    if (peerUid) {
                      const friend = chatStore.friends().find(f => f.uid === peerUid);
                      return friend?.online ? "在线" : "离线";
                    }
                    return "离线";
                  })())}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-1">
              <Show when={activeConv()?.type === "GROUP"}>
                <button
                  onClick={() => chatStore.setShowGroupPanel(!chatStore.showGroupPanel())}
                  class={`p-2 rounded-lg transition-colors ${
                    chatStore.showGroupPanel()
                      ? "bg-primary/10 text-primary"
                      : "text-text-secondary hover:text-text hover:bg-surface"
                  }`}
                >
                  <Info size={18} />
                </button>
              </Show>
              <button class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text">
                <Phone size={18} />
              </button>
              <button class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text">
                <Video size={18} />
              </button>
            </div>
          </div>

        {/* Messages */}
        <div
          ref={messagesContainerRef}
          class="flex-1 overflow-y-auto px-5 py-4"
          onScroll={handleScroll}
        >
          <Show when={chatStore.loading()}>
            <div class="text-center py-4">
              <div class="inline-block w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
            </div>
          </Show>
          <Show when={chatStore.hasMore() && !chatStore.loading()}>
            <div class="text-center py-2">
              <button
                onClick={() => chatStore.loadMoreMessages()}
                class="text-xs text-primary hover:text-primary-dark transition-colors flex items-center gap-1 mx-auto"
              >
                <ArrowUp size={12} />
                加载更多消息
              </button>
            </div>
          </Show>
          <For each={chatStore.messages()}>
            {(msg) => (
              <MessageBubble
                msg={msg}
                isMine={msg.sender === authStore.uid()}
                onRecall={() => handleRecall(msg)}
                onReply={() => {
                  setReplyTo(msg);
                  setEditingMsg(null);
                  inputRef?.focus();
                }}
                onEdit={() => handleEdit(msg)}
              />
            )}
          </For>
          <div ref={messagesEndRef} />
        </div>

        {/* Reply Preview */}
        <Show when={replyTo()}>
          <div class="px-5 py-2 bg-surface border-t border-border flex items-center gap-3">
            <Reply size={14} class="text-text-muted" />
            <div class="flex-1 min-w-0">
              <p class="text-xs text-primary font-medium">
                回复 {replyTo()?.sender_name}
              </p>
              <p class="text-xs text-text-muted truncate">
                {replyTo()?.content}
              </p>
            </div>
            <button
              onClick={() => setReplyTo(null)}
              class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted"
            >
              <X size={14} />
            </button>
          </div>
        </Show>

        {/* Edit Preview */}
        <Show when={editingMsg()}>
          <div class="px-5 py-2 bg-surface border-t border-border flex items-center gap-3">
            <Edit3 size={14} class="text-text-muted" />
            <div class="flex-1 min-w-0">
              <p class="text-xs text-primary font-medium">编辑消息</p>
              <p class="text-xs text-text-muted truncate">
                {editingMsg()?.content}
              </p>
            </div>
            <button
              onClick={() => {
                setEditingMsg(null);
                setInputText("");
              }}
              class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted"
            >
              <X size={14} />
            </button>
          </div>
        </Show>

        {/* Input */}
        <div class="px-5 py-3 border-t border-border bg-bg-secondary/30">
          <div class="flex items-end gap-2 bg-surface rounded-2xl border border-border focus-within:border-primary transition-colors p-2">
            <div class="flex items-center gap-1 px-1 pb-1">
              <button class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text">
                <Image size={18} />
              </button>
              <button class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text">
                <Paperclip size={18} />
              </button>
              <button class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text">
                <Smile size={18} />
              </button>
            </div>
            <textarea
              ref={inputRef}
              value={inputText()}
              onInput={(e) => {
                setInputText(e.currentTarget.value);
                handleInputTyping();
              }}
              onKeyDown={handleKeyDown}
              placeholder="输入消息..."
              rows={1}
              class="flex-1 bg-transparent resize-none text-sm text-text placeholder:text-text-muted focus:outline-none py-1.5 max-h-32"
            />
            <button
              onClick={handleSend}
              disabled={!inputText().trim()}
              class="p-2 bg-primary hover:bg-primary-dark disabled:opacity-40 disabled:cursor-not-allowed rounded-xl transition-all text-white"
            >
              <Send size={16} />
            </button>
          </div>
        </div>
      </Show>
      </div>

      <Show when={activeConv()?.type === "GROUP" && chatStore.showGroupPanel()}>
        <GroupInfoPanel />
      </Show>
    </div>
  );
}

function MessageBubble(props: {
  msg: MessageItem;
  isMine: boolean;
  onRecall: () => void;
  onReply: () => void;
  onEdit: () => void;
}) {
  const [showMenu, setShowMenu] = createSignal(false);
  const [showReadList, setShowReadList] = createSignal(false);

  const activeConv = () => {
    const id = chatStore.activeConvId();
    return chatStore.conversations().find((c) => c.conv_id === id);
  };

  const isGroup = () => activeConv()?.type === "GROUP";

  const readKey = () => `${props.msg.conv_id}:${props.msg.id}`;

  const readMembers = () => chatStore.readReceipts()[readKey()] || [];

  const readCount = () => {
    const local = props.msg.read_by?.length || 0;
    const server = readMembers().length;
    return Math.max(local, server);
  };

  const handleReadClick = () => {
    if (!isGroup() || !props.isMine) return;
    setShowReadList(!showReadList());
    if (!showReadList()) {
      chatStore.loadReadMembers(props.msg.conv_id, props.msg.id);
    }
  };

  let readListRef: HTMLDivElement | undefined;

  onMount(() => {
    const handler = (e: MouseEvent) => {
      if (readListRef && !readListRef.contains(e.target as Node)) {
        setShowReadList(false);
      }
    };
    document.addEventListener("click", handler);
    onCleanup(() => document.removeEventListener("click", handler));
  });

  const quotedMsg = () => {
    if (!props.msg.quote_msg_id) return null;
    return chatStore.messages().find((m) => m.id === props.msg.quote_msg_id);
  };

  const quoteContent = () => {
    const q = quotedMsg();
    if (!q) return props.msg.quote_content || null;
    if (q.recalled) return "消息已撤回";
    if (q.content_type === "text") return q.content;
    if (q.content_type?.startsWith("image/")) return "[图片]";
    return `[${q.content_type || "消息"}]`;
  };

  const quoteSender = () => {
    const q = quotedMsg();
    if (!q) return "";
    return q.sender_name || "";
  };

  return (
    <Show when={props.msg.recalled} fallback={
      <div
        id={`msg-${props.msg.id}`}
        data-msg-id={props.msg.id}
        class={`flex gap-2 mb-4 ${props.isMine ? "flex-row-reverse" : ""}`}
        onMouseEnter={() => setShowMenu(true)}
        onMouseLeave={() => setShowMenu(false)}
      >
        <Show when={!props.isMine}>
          <Avatar
            name={props.msg.sender_name}
            src={props.msg.sender_avatar}
            size="sm"
            class="mt-1"
          />
        </Show>

        <div class={`max-w-[70%] ${props.isMine ? "items-end" : "items-start"}`}>
          <Show when={!props.isMine}>
            <p class="text-xs text-text-muted mb-1 ml-1">
              {props.msg.sender_name}
            </p>
          </Show>

          {/* Quote */}
          <Show when={props.msg.quote_msg_id}>
            <div class="mb-1 px-3 py-1.5 bg-surface border-l-2 border-primary rounded-r-lg text-xs cursor-pointer hover:bg-surface-hover transition-colors" onClick={() => {
              const q = quotedMsg();
              if (q) {
                const el = document.getElementById(`msg-${q.id}`);
                el?.scrollIntoView({ behavior: "smooth", block: "center" });
              }
            }}>
              <Show when={quoteSender()}>
                <span class="font-medium text-primary">{quoteSender()}</span>
                <span class="text-text-muted mx-1">:</span>
              </Show>
              <span class="text-text-muted">{quoteContent() || "引用消息"}</span>
            </div>
          </Show>

          <div class="group relative">
            <div
              class={`px-4 py-2 rounded-2xl text-sm leading-relaxed break-words ${
                props.isMine
                  ? "bg-primary text-white rounded-br-md"
                  : "bg-surface text-text rounded-bl-md"
              }`}
            >
              {props.msg.content_type === "text" ? (
                props.msg.content
              ) : props.msg.content_type.startsWith("image/") ? (
                <img
                  src={props.msg.content}
                  alt="图片"
                  class="max-w-[300px] rounded-lg"
                />
              ) : (
                <div class="flex items-center gap-2">
                  <span>📎</span>
                  <span>{props.msg.content_type}</span>
                </div>
              )}
            </div>

            {/* Message actions */}
            <Show when={showMenu()}>
              <div
                class={`absolute top-0 ${
                  props.isMine ? "-left-[4.5rem]" : "-right-[4.5rem]"
                } flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity`}
              >
                <button
                  onClick={props.onReply}
                  class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                  title="回复"
                >
                  <Reply size={14} />
                </button>
                <Show when={props.isMine}>
                  <button
                    onClick={props.onEdit}
                    class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                    title="编辑"
                  >
                    <Edit3 size={14} />
                  </button>
                  <button
                    onClick={props.onRecall}
                    class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-danger"
                    title="撤回"
                  >
                    <X size={14} />
                  </button>
                </Show>
              </div>
            </Show>
          </div>

          <div
            class={`flex items-center gap-1 mt-0.5 ${
              props.isMine ? "justify-end mr-1" : "ml-1"
            }`}
          >
            <span class="text-xs text-text-muted">
              {formatTime(props.msg.created_at)}
            </span>
            <Show when={props.isMine}>
              {props.msg.edited ? (
                <span class="text-xs text-text-muted">已编辑</span>
              ) : (
                <CheckCheck size={12} class="text-text-muted" />
              )}
            </Show>
            <Show when={isGroup() && props.isMine && readCount() > 0}>
              <div class="relative">
                <button
                  onClick={handleReadClick}
                  class="flex items-center gap-0.5 text-xs text-text-muted hover:text-primary transition-colors cursor-pointer"
                  title="查看已读成员"
                >
                  <Eye size={12} />
                  <span>{readCount()}</span>
                </button>
                <Show when={showReadList()}>
                  <div ref={readListRef} class="absolute bottom-full right-0 mb-1 bg-surface border border-border rounded-xl shadow-lg z-50 py-2 min-w-[180px] max-h-[200px] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
                    <p class="px-3 py-1 text-xs text-text-muted font-medium">已读 {readCount()} 人</p>
                    <For each={readMembers()}>
                      {(member) => (
                        <div class="flex items-center gap-2 px-3 py-1.5 hover:bg-bg transition-colors">
                          <Avatar name={member.name} src={member.avatar} size="sm" />
                          <span class="text-sm text-text">{member.name}</span>
                        </div>
                      )}
                    </For>
                    <Show when={props.msg.read_by && props.msg.read_by.length > 0 && readMembers().length === 0}>
                      <div class="px-3 py-2 text-xs text-text-muted text-center">加载中...</div>
                    </Show>
                  </div>
                </Show>
              </div>
            </Show>
          </div>
        </div>
      </div>
    }>
      <div id={`msg-${props.msg.id}`} class="flex justify-center mb-3">
        <span class="text-xs text-text-muted bg-surface/50 px-3 py-1 rounded-full">
          {props.isMine ? "你" : props.msg.sender_name || "用户"}撤回了一条消息
        </span>
      </div>
    </Show>
  );
}