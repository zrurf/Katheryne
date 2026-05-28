import { For, Show, createSignal, onCleanup, onMount, createEffect, createMemo } from "solid-js";
import { marked } from "marked";
import { chatStore } from "../../stores/chat";
import { authStore } from "../../stores/auth";
import { api } from "../../services/api";
import { downloadFile } from "../../services/download";
import { formatMessageSnippet } from "../../lib/utils";
import { getServerApiBase } from "../../services/config";
import { Avatar } from "../ui/avatar";
import { ImageViewer } from "../ui/image-viewer";
import { GroupInfoPanel } from "./GroupInfoPanel";
import { EmojiPicker } from "./EmojiPicker";
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
  ArrowDown,
  Eye,
  Loader2,
  Sparkles,
  Languages,
  FileText,
  Lightbulb,
} from "lucide-solid";

interface PendingUpload {
  id: string;
  convId: string;
  fileName: string;
  fileType: "image" | "file";
  file: File; // retained for retry
  fileSize: number;
  progress: number; // 0-100
  status: "uploading" | "error";
  errorMsg?: string;
  previewUrl?: string; // object URL for image preview
  replyToId?: string;
}

export function ChatView() {
  const [inputText, setInputText] = createSignal("");
  const [replyTo, setReplyTo] = createSignal<MessageItem | null>(null);
  const [editingMsg, setEditingMsg] = createSignal<MessageItem | null>(null);
  const [showEmoji, setShowEmoji] = createSignal(false);
  const [uploading, setUploading] = createSignal(false);
  const [pendingUploads, setPendingUploads] = createSignal<PendingUpload[]>([]);
  // @ mention state
  const [showMention, setShowMention] = createSignal(false);
  const [mentionQuery, setMentionQuery] = createSignal("");
  const [mentionItems, setMentionItems] = createSignal<{ id: string; name: string; type: "member" | "bot" }[]>([]);
  const [mentionIndex, setMentionIndex] = createSignal(0);
  const [mentionStartPos, setMentionStartPos] = createSignal(0);
  const [mentionBots, setMentionBots] = createSignal<{ id: string; name: string }[]>([]);
  // Bot interaction state
  const [botActioning, setBotActioning] = createSignal<string | null>(null); // "summarize" | "suggest" | "translate"
  const [botResult, setBotResult] = createSignal<{
    type: string;
    title: string;
    content: string;
    items?: string[];
  } | null>(null);
  const [translatingMsgId, setTranslatingMsgId] = createSignal<string | null>(null);
  const [translateMap, setTranslateMap] = createSignal<Record<string, string>>({});
  const [imageViewer, setImageViewer] = createSignal<{ images: string[]; index: number } | null>(null);
  const [showBotPanel, setShowBotPanel] = createSignal(false);
  let messagesEndRef: HTMLDivElement | undefined;
  let messagesContainerRef: HTMLDivElement | undefined;
  let inputRef: HTMLDivElement | undefined;
  let typingTimer: ReturnType<typeof setTimeout> | null = null;
  let isTyping = false;
  let readTimer: ReturnType<typeof setTimeout> | null = null;
  let lastSubmittedEndMsgId = "";
  // 未读消息悬浮标签
  const [badgeCount, setBadgeCount] = createSignal(0);
  const [badgeFirstMsgId, setBadgeFirstMsgId] = createSignal("");
  let prevMsgCount = 0;
  let imageInputRef: HTMLInputElement | undefined;
  let fileInputRef: HTMLInputElement | undefined;

  const activeConv = () => {
    const id = chatStore.activeConvId();
    return chatStore.conversations().find((c) => c.conv_id === id);
  };

  // Time gap threshold in ms — insert a time separator when messages are >5min apart
  const TIME_GAP_THRESHOLD = 5 * 60 * 1000;

  const messageList = createMemo(() => {
    const msgs = chatStore.messages();
    const result: ({ _tag: "message"; msg: MessageItem } | { _tag: "separator"; timestamp: number })[] = [];

    for (let i = 0; i < msgs.length; i++) {
      const msg = msgs[i];
      const prev = i > 0 ? msgs[i - 1] : null;

      // Insert time separator before the first message or when gap exceeds threshold
      if (!prev || (msg.created_at - prev.created_at) > TIME_GAP_THRESHOLD) {
        result.push({ _tag: "separator", timestamp: msg.created_at });
      }

      result.push({ _tag: "message", msg });
    }

    return result;
  });

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

  const isNearBottom = () => {
    const container = messagesContainerRef;
    if (!container) return true;
    return container.scrollHeight - container.scrollTop - container.clientHeight < 100;
  };

  createEffect(() => {
    const msgs = chatStore.messages();
    if (msgs.length > 0) {
      // 如果用户在底部或之前就在底部，自动滚到底部并清除标签
      if (isNearBottom() || prevMsgCount === 0) {
        scrollToBottom();
        setBadgeCount(0);
        setBadgeFirstMsgId("");
        setTimeout(() => trackVisibleMessages(), 100);
      } else if (msgs.length > prevMsgCount) {
        // 有新消息但用户不在底部，递增标签计数
        setBadgeCount((c) => c + (msgs.length - prevMsgCount));
        if (!badgeFirstMsgId()) {
          // 记录第一条新消息
          setBadgeFirstMsgId(String(msgs[prevMsgCount]?.id || ""));
        }
      }
    }
    prevMsgCount = msgs.length;
  });

  createEffect(() => {
    chatStore.activeConvId();
    lastSubmittedEndMsgId = "";
    prevMsgCount = 0;
    setBadgeCount(0);
    setBadgeFirstMsgId("");
  });

  // 会话首次进入后，设置未读标签
  createEffect(() => {
    const count = chatStore.activeConvUnreadCount();
    const firstId = chatStore.firstUnreadMsgId();
    if (count > 0 && firstId) {
      setBadgeCount(count);
      setBadgeFirstMsgId(firstId);
    }
  });

  const handleScroll = () => {
    const container = messagesContainerRef;
    if (!container) return;
    if (container.scrollTop < 50 && chatStore.hasMore() && !chatStore.loading()) {
      chatStore.loadMoreMessages();
    }
    trackVisibleMessages();
    // 用户滚动到底部则清除标签
    if (isNearBottom()) {
      setBadgeCount(0);
      setBadgeFirstMsgId("");
    }
  };

  const handleBadgeClick = () => {
    const msgId = badgeFirstMsgId();
    if (msgId) {
      const el = document.getElementById(`msg-${msgId}`);
      el?.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    setBadgeCount(0);
    setBadgeFirstMsgId("");
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

  // ============ ContentEditable helpers ============
  /** Extract raw text (with @[type:id:name] tags) from contentEditable DOM */
  function getEditableText(div: HTMLDivElement): string {
    let result = "";
    const walk = (node: Node) => {
      if (node.nodeType === Node.TEXT_NODE) {
        result += node.textContent || "";
      } else if (node.nodeType === Node.ELEMENT_NODE) {
        const el = node as HTMLElement;
        if (el.classList.contains("mention-pill")) {
          const type = el.dataset.type || "member";
          const id = el.dataset.id || "";
          const name = el.dataset.name || "";
          result += `@[${type}:${id}:${name}]`;
        } else if (el.tagName === "BR") {
          result += "\n";
        } else {
          for (const child of el.childNodes) {
            walk(child);
          }
        }
      }
    };
    for (const child of div.childNodes) {
      walk(child);
    }
    // Trim trailing newline from &nbsp;<br> pattern
    return result.replace(/\n$/, "");
  }

  /** Sync inputText signal from contentEditable DOM */
  function syncInputFromEditable() {
    const div = inputRef;
    if (!div) return;
    const text = getEditableText(div);
    setInputText(text);
  }

  /** Set contentEditable content from raw text */
  function setEditableContent(div: HTMLDivElement, text: string) {
    let html = "";
    let lastIndex = 0;
    const pattern = /@\[([a-z]+):(\d+):([^\]]+)\]/g;
    let match;
    while ((match = pattern.exec(text)) !== null) {
      if (match.index > lastIndex) {
        html += escapeHtml(text.slice(lastIndex, match.index));
      }
      const mentionType = match[1];
      const id = match[2];
      const name = match[3];
      html += `<span class="mention-pill ${mentionType === "bot" ? "mention-bot" : "mention-user"}" contenteditable="false" data-type="${mentionType}" data-id="${id}" data-name="${escapeHtml(name)}">@${escapeHtml(name)}</span>`;
      lastIndex = match.index + match[0].length;
    }
    if (lastIndex < text.length) {
      html += escapeHtml(text.slice(lastIndex));
    }
    div.innerHTML = html || "";
    // Move cursor to end
    const range = document.createRange();
    range.selectNodeContents(div);
    range.collapse(false);
    const sel = window.getSelection();
    sel?.removeAllRanges();
    sel?.addRange(range);
  }

  /** Get text before cursor for @mention detection */
  function getTextBeforeCursor(): { text: string; node: Node | null; offset: number } {
    const sel = window.getSelection();
    if (!sel || !sel.rangeCount) return { text: "", node: null, offset: 0 };
    const range = sel.getRangeAt(0);
    const node = sel.anchorNode;
    const offset = sel.anchorOffset;
    if (node && node.nodeType === Node.TEXT_NODE) {
      return { text: node.textContent?.substring(0, offset) || "", node, offset };
    }
    return { text: "", node: null, offset: 0 };
  }

  /** Insert a mention pill at the current cursor (replacing the @pattern range) */
  function insertMentionAtCursor(type: string, id: string, name: string) {
    const div = inputRef;
    if (!div) return;
    const sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;

    // We need to delete from @ to cursor, then insert the pill
    const range = sel.getRangeAt(0);
    const atPos = mentionStartPos();

    // Walk backwards from cursor to find and delete @query text
    let textNode = sel.anchorNode;
    if (textNode && textNode.nodeType === Node.TEXT_NODE) {
      const before = textNode.textContent?.substring(0, sel.anchorOffset) || "";
      const localAt = before.lastIndexOf("@");
      if (localAt >= 0) {
        range.setStart(textNode, localAt);
        range.setEnd(textNode, sel.anchorOffset);
        range.deleteContents();
      }
    }

    // Create and insert mention span
    const span = document.createElement("span");
    span.className = `mention-pill ${type === "bot" ? "mention-bot" : "mention-user"}`;
    span.contentEditable = "false";
    span.dataset.type = type;
    span.dataset.id = id;
    span.dataset.name = name;
    span.textContent = `@${name}`;

    range.insertNode(span);

    // Add a space after, then move cursor after space
    const space = document.createTextNode("\u00A0"); // non-breaking space
    range.setStartAfter(span);
    range.collapse(true);
    range.insertNode(space);

    range.setStartAfter(space);
    range.collapse(true);
    sel.removeAllRanges();
    sel.addRange(range);

    div.focus();
    setShowMention(false);
    syncInputFromEditable();
  }

  // Replace the old overlay rendering with contentEditable sync
  createEffect(() => {
    const text = inputText();
    const div = inputRef;
    if (!div) return;
    // Only update when the signal changes externally (reply, edit, emoji)
    // Input events update the signal, so we skip re-rendering on those
    const currentText = getEditableText(div);
    if (currentText !== text) {
      // Store cursor position before rebuilding
      const sel = window.getSelection();
      let cursorOffset = 0;
      if (sel && sel.rangeCount && div.contains(sel.anchorNode)) {
        // Try to preserve approximate cursor position
        const range = sel.getRangeAt(0);
        const preRange = document.createRange();
        preRange.selectNodeContents(div);
        preRange.setEnd(range.endContainer, range.endOffset);
        cursorOffset = preRange.toString().length;
      }
      setEditableContent(div, text);
      // Restore cursor position
      if (cursorOffset > 0) {
        restoreCursor(div, cursorOffset);
      }
    }
  });

  function restoreCursor(div: HTMLDivElement, offset: number) {
    const sel = window.getSelection();
    if (!sel) return;
    const range = document.createRange();
    let count = 0;
    const walk = (node: Node): boolean => {
      if (node.nodeType === Node.TEXT_NODE) {
        const len = node.textContent?.length || 0;
        if (count + len >= offset) {
          range.setStart(node, offset - count);
          range.collapse(true);
          return true;
        }
        count += len;
      } else if (node.nodeType === Node.ELEMENT_NODE) {
        const el = node as HTMLElement;
        if (el.classList.contains("mention-pill")) {
          count += 1; // approximate
        }
        for (const child of el.childNodes) {
          if (walk(child)) return true;
        }
      }
      return false;
    };
    for (const child of div.childNodes) {
      if (walk(child)) break;
    }
    sel.removeAllRanges();
    sel.addRange(range);
  }

  const handleSend = () => {
    const text = inputText().trim();
    if (!text) return;

    const conv = activeConv();
    if (!conv) return;

    // Extract mentions: @[type:uid:nickname]
    const mentionPattern = /@\[([a-z]+):(\d+):([^\]]+)\]/g;
    const mentions: { type: string; id: string; name: string }[] = [];
    let match;
    while ((match = mentionPattern.exec(text)) !== null) {
      mentions.push({ type: match[1], id: match[2], name: match[3] });
    }
    const extra = mentions.length > 0 ? JSON.stringify({ mentions }) : undefined;

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
        replyTo()?.id,
        extra
      );
    }
    setInputText("");
    setReplyTo(null);
    scrollToBottom();
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (showMention()) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        const items = mentionItems();
        setMentionIndex((items.length + mentionIndex() + 1) % items.length);
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        const items = mentionItems();
        setMentionIndex((items.length + mentionIndex() - 1) % items.length);
        return;
      }
      if (e.key === "Enter" || e.key === "Tab") {
        e.preventDefault();
        const items = mentionItems();
        if (items.length > 0) {
          selectMention(items[mentionIndex()]);
        }
        return;
      }
      if (e.key === "Escape") {
        setShowMention(false);
        return;
      }
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleEmojiSelect = (emoji: string) => {
    const div = inputRef;
    if (!div) { setInputText((prev) => prev + emoji); setShowEmoji(false); return; }
    const sel = window.getSelection();
    if (sel && sel.rangeCount && div.contains(sel.anchorNode)) {
      const range = sel.getRangeAt(0);
      range.deleteContents();
      const textNode = document.createTextNode(emoji);
      range.insertNode(textNode);
      range.setStartAfter(textNode);
      range.collapse(true);
      sel.removeAllRanges();
      sel.addRange(range);
    }
    div.focus();
    syncInputFromEditable();
    setShowEmoji(false);
  };

  // Shared upload helper: auto-selects chunked upload for large files (>5MB)
  // and enforces the server-side max file size limit.
  const doUpload = async (
    file: File,
    onProgress: (pct: number) => void,
  ): Promise<{ url: string; oss_index: string; index_id: string; filename: string }> => {
    // Fetch server config for size limit
    let maxSize = 100 * 1024 * 1024; // fallback 100MB
    try {
      const cfg = await api.oss.getConfig();
      maxSize = cfg.max_file_size;
    } catch { /* use fallback */ }

    if (file.size > maxSize) {
      throw new Error(`文件 ${file.name} 大小 ${formatFileSize(file.size)} 超过限制 ${formatFileSize(maxSize)}`);
    }

    const CHUNKED_THRESHOLD = 5 * 1024 * 1024; // 5MB — use chunked above this
    if (file.size > CHUNKED_THRESHOLD) {
      return api.oss.chunkedUpload(file, onProgress);
    }
    return api.oss.uploadWithProgress(file, onProgress);
  };

  const retryUpload = async (upload: PendingUpload) => {
    setPendingUploads((prev) =>
      prev.map((p) =>
        p.id === upload.id ? { ...p, status: "uploading" as const, progress: 0, errorMsg: undefined } : p
      )
    );
    try {
      const uploadResp = await doUpload(upload.file, (pct) => {
        setPendingUploads((prev) =>
          prev.map((p) => (p.id === upload.id ? { ...p, progress: pct } : p))
        );
      });

      if (upload.fileType === "image") {
        const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
        await chatStore.sendMessage(
          upload.convId,
          "",
          "image",
          proxyPath,
          upload.file.type,
          upload.replyToId
        );
      } else {
        const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
        const content = JSON.stringify({
          name: uploadResp.filename,
          size: upload.fileSize,
          url: proxyPath,
          oss_index: uploadResp.oss_index,
          index_id: uploadResp.index_id,
        });
        await chatStore.sendMessage(
          upload.convId,
          "",
          "file",
          content,
          upload.file.type,
          upload.replyToId
        );
      }

      setPendingUploads((prev) => prev.filter((p) => p.id !== upload.id));
      scrollToBottom();
    } catch (err) {
      console.error("Retry upload failed:", err);
      setPendingUploads((prev) =>
        prev.map((p) =>
          p.id === upload.id
            ? { ...p, status: "error", errorMsg: (err as Error).message || "上传失败" }
            : p
        )
      );
    }
  };

  const dismissUpload = (uploadId: string) => {
    setPendingUploads((prev) => {
      const target = prev.find((p) => p.id === uploadId);
      if (target?.previewUrl) {
        URL.revokeObjectURL(target.previewUrl);
      }
      return prev.filter((p) => p.id !== uploadId);
    });
  };

  const handleImageUpload = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    const conv = activeConv();
    if (!conv) return;

    if (!file.type.startsWith("image/")) {
      alert("请选择图片文件");
      return;
    }

    const uploadId = `upload_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const previewUrl = URL.createObjectURL(file);

    // Immediately show pending message bubble with progress
    setPendingUploads((prev) => [
      ...prev,
      {
        id: uploadId,
        convId: conv.conv_id,
        fileName: file.name,
        fileType: "image",
        file: file,
        fileSize: file.size,
        progress: 0,
        status: "uploading",
        previewUrl,
        replyToId: replyTo()?.id,
      },
    ]);
    scrollToBottom();

    try {
      const uploadResp = await api.oss.uploadWithProgress(file, (pct) => {
        setPendingUploads((prev) =>
          prev.map((p) => (p.id === uploadId ? { ...p, progress: pct } : p))
        );
      });

      // Upload done — send the real message
      const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
      await chatStore.sendMessage(
        conv.conv_id,
        "",
        "image",
        proxyPath,
        file.type,
        replyTo()?.id
      );
      setReplyTo(null);

      // Remove pending upload
      setPendingUploads((prev) => prev.filter((p) => p.id !== uploadId));
      scrollToBottom();
    } catch (err) {
      console.error("Image upload failed:", err);
      setPendingUploads((prev) =>
        prev.map((p) =>
          p.id === uploadId
            ? { ...p, status: "error", errorMsg: (err as Error).message || "上传失败" }
            : p
        )
      );
    } finally {
      if (input) input.value = "";
    }
  };

  const handleFileUpload = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    const conv = activeConv();
    if (!conv) return;

    const uploadId = `upload_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;

    // Immediately show pending message bubble with progress
    setPendingUploads((prev) => [
      ...prev,
      {
        id: uploadId,
        convId: conv.conv_id,
        fileName: file.name,
        fileType: "file",
        file: file,
        fileSize: file.size,
        progress: 0,
        status: "uploading",
        replyToId: replyTo()?.id,
      },
    ]);
    scrollToBottom();

    try {
      const uploadResp = await doUpload(file, (pct) => {
        setPendingUploads((prev) =>
          prev.map((p) => (p.id === uploadId ? { ...p, progress: pct } : p))
        );
      });

      // Upload done — send the real message
      const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
      const content = JSON.stringify({
        name: uploadResp.filename,
        size: file.size,
        url: proxyPath,
        oss_index: uploadResp.oss_index,
        index_id: uploadResp.index_id,
      });
      await chatStore.sendMessage(
        conv.conv_id,
        "",
        "file",
        content,
        file.type,
        replyTo()?.id
      );
      setReplyTo(null);

      // Remove pending upload
      setPendingUploads((prev) => prev.filter((p) => p.id !== uploadId));
      scrollToBottom();
    } catch (err) {
      console.error("File upload failed:", err);
      setPendingUploads((prev) =>
        prev.map((p) =>
          p.id === uploadId
            ? { ...p, status: "error", errorMsg: (err as Error).message || "上传失败" }
            : p
        )
      );
    } finally {
      if (input) input.value = "";
    }
  };

  // @ mention helpers
  const loadMentionBots = async () => {
    if (mentionBots().length > 0) return;
    const conv = activeConv();
    if (!conv || conv.type !== "GROUP") return;
    try {
      const resp = await api.bot.getConvBots(conv.conv_id);
      setMentionBots(
        resp.list.map((b) => ({ id: String(b.bot_id), name: b.name }))
      );
    } catch {
      // ignore
    }
  };

  const getMentionCandidates = (query: string) => {
    const q = query.toLowerCase();
    const members = chatStore.groupMembers()
      .filter((m) => {
        const mName = m.name || "";
        const mNick = m.nick || "";
        return mName.toLowerCase().includes(q) || mNick.toLowerCase().includes(q);
      })
      .map((m) => ({
        id: String(m.uid),
        name: (m.nick || m.name || "") as string,
        type: "member" as const,
      }));

    const bots = mentionBots()
      .filter((b) => b.name.toLowerCase().includes(q))
      .map((b) => ({
        id: b.id,
        name: b.name,
        type: "bot" as const,
      }));

    return [...members, ...bots].slice(0, 10);
  };

  const handleMentionInput = (_value: string /* unused - we read from DOM */, cursorPos: number) => {
    const before = getTextBeforeCursor();
    const beforeText = before.text;
    const atPos = beforeText.lastIndexOf("@");
    if (atPos === -1 || (atPos > 0 && beforeText[atPos - 1] !== " " && beforeText[atPos - 1] !== "\n")) {
      setShowMention(false);
      return;
    }
    const afterAt = beforeText.substring(atPos + 1);
    if (afterAt.includes(" ")) {
      setShowMention(false);
      return;
    }
    setMentionQuery(afterAt);
    setMentionStartPos(atPos);
    setMentionIndex(0);
    setMentionItems(getMentionCandidates(afterAt));
    setShowMention(true);
    loadMentionBots();
  };

  const selectMention = (item: { id: string; name: string; type: string }) => {
    insertMentionAtCursor(item.type, item.id, item.name);
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

  // Bot interaction handlers
  const handleSummarize = async () => {
    const conv = activeConv();
    if (!conv) return;
    setBotActioning("summarize");
    setBotResult(null);
    try {
      // 提交总结请求，获取 ticket
      const ticketResp = await api.bot.summarize(conv.conv_id);

      // 轮询查询结果（最多轮询 2 分钟）
      const maxAttempts = 80;
      let attempts = 0;
      const poll = async (): Promise<void> => {
        if (attempts++ >= maxAttempts) {
          setBotResult({
            type: "error",
            title: "总结超时",
            content: "总结请求超时，请稍后再试",
          });
          setShowBotPanel(true);
          setBotActioning(null);
          return;
        }
        const resultResp = await api.bot.summarizeResult(ticketResp.ticket);
        if (resultResp.status === "pending" || resultResp.status === "processing") {
          // 继续轮询
          await new Promise((r) => setTimeout(r, 1500));
          return poll();
        }
        if (resultResp.status === "error" || resultResp.error) {
          setBotResult({
            type: "error",
            title: "总结失败",
            content: resultResp.error || "请稍后再试",
          });
          setShowBotPanel(true);
          setBotActioning(null);
          return;
        }
        if (resultResp.status === "completed" && resultResp.result) {
          setBotResult({
            type: "summarize",
            title: "对话总结",
            content: resultResp.result.summary,
            items: [...resultResp.result.key_points, ...resultResp.result.action_items.map(a => `📋 ${a}`)],
          });
          setShowBotPanel(true);
          setBotActioning(null);
        }
      };
      await poll();
    } catch (err) {
      console.error("Summarize failed:", err);
      setBotResult({
        type: "error",
        title: "总结失败",
        content: "请稍后再试",
      });
      setShowBotPanel(true);
      setBotActioning(null);
    }
  };

  const handleSuggestReplies = async () => {
    const conv = activeConv();
    if (!conv) return;
    setBotActioning("suggest");
    setBotResult(null);
    try {
      const resp = await api.bot.suggestReplies(conv.conv_id);
      setBotResult({
        type: "suggest",
        title: "回复建议",
        content: "",
        items: resp.suggestions,
      });
      setShowBotPanel(true);
    } catch (err) {
      console.error("Suggest failed:", err);
      setBotResult({
        type: "error",
        title: "获取建议失败",
        content: "请稍后再试",
      });
      setShowBotPanel(true);
    } finally {
      setBotActioning(null);
    }
  };

  const handleTranslateMsg = async (msg: MessageItem) => {
    if (msg.content_type !== "text") return;
    setTranslatingMsgId(msg.id);
    try {
      const resp = await api.bot.translate(msg.content, "zh", "auto");
      setTranslateMap(prev => ({ ...prev, [msg.id]: resp.text }));
    } catch (err) {
      console.error("Translate failed:", err);
    } finally {
      setTranslatingMsgId(null);
    }
  };

  const openImageViewer = (url: string) => {
    // Collect all image URLs from current messages
    const allImages: string[] = [];
    for (const item of messageList()) {
      if (item._tag === "message" && item.msg.content_type?.startsWith("image/")) {
        allImages.push(item.msg.content);
      }
    }
    const idx = Math.max(0, allImages.indexOf(url));
    setImageViewer({ images: allImages, index: idx });
  };

  const useSuggestion = (text: string) => {
    setInputText(text);
    setShowBotPanel(false);
    inputRef?.focus();
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
                  onClick={() => chatStore.toggleGroupPanel()}
                  class={`p-2 rounded-lg transition-colors ${
                    chatStore.showGroupPanel()
                      ? "bg-primary/10 text-primary"
                      : "text-text-secondary hover:text-text hover:bg-surface"
                  }`}
                >
                  <Info size={18} />
                </button>
              </Show>
              {/* Bot: Summarize */}
              <button
                onClick={handleSummarize}
                disabled={botActioning() !== null}
                class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text disabled:opacity-40"
                title="总结对话"
              >
                {botActioning() === "summarize" ? (
                  <Loader2 size={18} class="animate-spin" />
                ) : (
                  <FileText size={18} />
                )}
              </button>
              {/* Bot: Suggest Replies */}
              <button
                onClick={handleSuggestReplies}
                disabled={botActioning() !== null}
                class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text disabled:opacity-40"
                title="生成回复建议"
              >
                {botActioning() === "suggest" ? (
                  <Loader2 size={18} class="animate-spin" />
                ) : (
                  <Lightbulb size={18} />
                )}
              </button>
              <button class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text">
                <Phone size={18} />
              </button>
              <button class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text">
                <Video size={18} />
              </button>
            </div>
          </div>

        {/* Bot Result Panel */}
        <Show when={showBotPanel() && botResult()}>
          <div class="px-5 py-3 bg-surface border-b border-border">
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <Sparkles size={16} class="text-primary" />
                <span class="text-sm font-semibold text-text">{botResult()!.title}</span>
              </div>
              <button
                onClick={() => setShowBotPanel(false)}
                class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted"
              >
                <X size={14} />
              </button>
            </div>
            <Show when={botResult()!.type === "error"}>
              <p class="text-sm text-danger">{botResult()!.content}</p>
            </Show>
            <Show when={botResult()!.type === "summarize"}>
              <p class="text-sm text-text mb-2">{botResult()!.content}</p>
              <Show when={botResult()!.items && botResult()!.items!.length > 0}>
                <div class="flex flex-wrap gap-1.5">
                  <For each={botResult()!.items}>
                    {(item) => (
                      <span class="inline-block px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-full">
                        {item}
                      </span>
                    )}
                  </For>
                </div>
              </Show>
            </Show>
            <Show when={botResult()!.type === "suggest"}>
              <Show when={botResult()!.items && botResult()!.items!.length > 0}>
                <div class="space-y-1.5">
                  <For each={botResult()!.items}>
                    {(item) => (
                      <button
                        onClick={() => useSuggestion(item)}
                        class="w-full text-left px-3 py-2 bg-bg-secondary rounded-lg text-sm text-text hover:bg-primary/5 hover:text-primary transition-colors"
                      >
                        {item}
                      </button>
                    )}
                  </For>
                </div>
              </Show>
            </Show>
          </div>
        </Show>

        {/* Messages */}
        <div
          ref={messagesContainerRef}
          class="flex-1 overflow-y-auto px-5 py-4 relative"
          onScroll={handleScroll}
        >
          {/* 未读消息悬浮标签 */}
          <Show when={badgeCount() > 0}>
            <div class="sticky top-4 z-10 flex justify-center pointer-events-none">
              <button
                onClick={handleBadgeClick}
                class="pointer-events-auto flex items-center gap-1.5 px-3 py-1.5 bg-primary text-white text-xs font-medium rounded-full shadow-lg hover:bg-primary-dark transition-colors"
              >
                {badgeCount()} 条新消息
                <ArrowDown size={12} />
              </button>
            </div>
          </Show>
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
          <For each={messageList()}>
            {(item) => {
              if (item._tag === "separator") {
                return (
                  <div class="flex items-center justify-center my-3">
                    <span class="px-3 py-0.5 text-xs text-text-muted bg-surface/80 rounded-full select-none">
                      {formatTime(item.timestamp)}
                    </span>
                  </div>
                );
              }
              const msg = item.msg;
              return (
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
                  onTranslate={() => handleTranslateMsg(msg)}
                  isTranslating={translatingMsgId() === msg.id}
                  translateText={translateMap()[msg.id]}
                  onImageClick={openImageViewer}
                />
              );
            }}
          </For>
          <For each={pendingUploads()}>
            {(pending) => (
              <div class="mb-3">
                <PendingUploadBubble
                  upload={pending}
                  onRetry={retryUpload}
                  onDismiss={dismissUpload}
                />
              </div>
            )}
          </For>
          <div ref={messagesEndRef} />
        </div>

        {/* Reply Preview */}
        <Show when={replyTo()}>
          <div class="px-5 py-2 bg-surface border-t border-border flex items-center gap-3">
            <Reply size={14} class="text-text-muted shrink-0" />
            <div class="flex-1 min-w-0 overflow-hidden">
              <p class="text-xs text-primary font-medium truncate">
                回复 {replyTo()?.sender_name}
              </p>
              <p class="text-xs text-text-muted line-clamp-2 break-all">
                {formatMessageSnippet(
                  replyTo()?.content || "",
                  replyTo()?.content_type,
                  replyTo()?.type
                )}
              </p>
            </div>
            <button
              onClick={() => setReplyTo(null)}
              class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted shrink-0"
            >
              <X size={14} />
            </button>
          </div>
        </Show>

        {/* Edit Preview */}
        <Show when={editingMsg()}>
          <div class="px-5 py-2 bg-surface border-t border-border flex items-center gap-3 min-w-0">
            <div class="flex-1 min-w-0 overflow-hidden">
              <p class="text-xs text-primary font-medium">编辑消息</p>
              <p class="text-xs text-text-muted line-clamp-2 break-all">
                {editingMsg()?.content}
              </p>
            </div>
            <button
              onClick={() => {
                setEditingMsg(null);
                setInputText("");
              }}
              class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted shrink-0"
            >
              <X size={14} />
            </button>
          </div>
        </Show>

        {/* Input */}
        <div class="px-5 py-3 border-t border-border bg-bg-secondary/30">
          <div class="relative">
            {/* @ Mention dropdown */}
            <Show when={showMention() && mentionItems().length > 0}>
              <div class="absolute bottom-full left-0 right-0 mb-2 bg-surface border border-border rounded-xl shadow-lg z-50 max-h-48 overflow-y-auto">
                <p class="px-3 py-1.5 text-xs text-text-muted font-medium">选择要 @ 的用户</p>
                <For each={mentionItems()}>
                  {(item, idx) => (
                    <button
                      onClick={() => selectMention(item)}
                      class={`w-full flex items-center gap-2 px-3 py-2 text-sm text-left transition-colors ${
                        idx() === mentionIndex()
                          ? "bg-primary/10 text-primary"
                          : "text-text hover:bg-surface-hover"
                      }`}
                    >
                      <span class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-xs text-primary shrink-0">
                        {item.type === "bot" ? "B" : item.name.charAt(0)}
                      </span>
                      <span>{item.name}</span>
                    </button>
                  )}
                </For>
              </div>
            </Show>
          <div class="flex items-end gap-2 bg-surface rounded-2xl border border-border focus-within:border-primary transition-colors p-2">
            <div class="flex items-center gap-1 px-1 pb-1 relative">
              {/* Hidden file inputs */}
              <input
                ref={imageInputRef}
                type="file"
                accept="image/*"
                class="hidden"
                onChange={handleImageUpload}
              />
              <input
                ref={fileInputRef}
                type="file"
                class="hidden"
                onChange={handleFileUpload}
              />
              <button
                onClick={() => imageInputRef?.click()}
                disabled={uploading()}
                class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text disabled:opacity-40"
                title="发送图片"
              >
                {uploading() ? <Loader2 size={18} class="animate-spin" /> : <Image size={18} />}
              </button>
              <button
                onClick={() => fileInputRef?.click()}
                disabled={uploading()}
                class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text disabled:opacity-40"
                title="发送文件"
              >
                <Paperclip size={18} />
              </button>
              <button
                onClick={() => setShowEmoji(!showEmoji())}
                class="p-1.5 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text"
                title="表情"
              >
                <Smile size={18} />
              </button>
              {/* Emoji Picker */}
              <Show when={showEmoji()}>
                <EmojiPicker
                  onSelect={handleEmojiSelect}
                  onClose={() => setShowEmoji(false)}
                />
              </Show>
            </div>
            <div class="relative flex-1 min-w-0">
              <div
                ref={inputRef}
                contentEditable={true}
                role="textbox"
                aria-multiline="true"
                data-placeholder="输入消息..."
                onInput={() => {
                  syncInputFromEditable();
                  handleInputTyping();
                }}
                onKeyDown={handleKeyDown}
                onKeyUp={() => {
                  handleMentionInput("", 0);
                }}
                class="relative w-full bg-transparent resize-none text-sm text-text placeholder:text-text-muted focus:outline-none py-1.5 max-h-32 overflow-y-auto whitespace-pre-wrap break-words empty:before:content-[attr(data-placeholder)] empty:before:text-text-muted"
              />
            </div>
            <button
              onClick={handleSend}
              disabled={!inputText().trim() || uploading()}
              class="p-2 bg-primary hover:bg-primary-dark disabled:opacity-40 disabled:cursor-not-allowed rounded-xl transition-all text-white"
            >
              <Send size={16} />
            </button>
          </div>
          </div>
        </div>
      </Show>
      </div>

      <Show when={activeConv()?.type === "GROUP" && chatStore.showGroupPanel()}>
        <GroupInfoPanel />
      </Show>

      {/* Image Viewer */}
      <Show when={imageViewer()}>
        <ImageViewer
          images={imageViewer()!.images}
          initialIndex={imageViewer()!.index}
          onClose={() => setImageViewer(null)}
        />
      </Show>
    </div>
  );
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/\n/g, "<br>");
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

/**
 * Renders message text content, parsing @[type:uid:nickname] mentions into styled tags.
 * type: "member" or "bot"
 */
function MentionText(props: { content: string; isMine: boolean }) {
  const parts: unknown[] = [];
  let lastIndex = 0;
  const pattern = /@\[([a-z]+):(\d+):([^\]]+)\]/g;
  let match;
  let idx = 0;
  while ((match = pattern.exec(props.content)) !== null) {
    if (match.index > lastIndex) {
      parts.push(
        <>{props.content.slice(lastIndex, match.index)}</>
      );
    }
    const mentionType = match[1]; // "member" or "bot"
    const uid = match[2];
    const name = match[3];
    const isBot = mentionType === "bot";
    parts.push(
      <span
        class={`inline-flex items-center gap-0.5 font-medium cursor-pointer hover:underline ${
          props.isMine
            ? isBot ? "text-white/90 bg-white/30" : "text-white/90 bg-white/20"
            : isBot ? "text-info bg-info/10" : "text-primary bg-primary/10"
        } rounded px-1 py-px`}
        data-mention-type={mentionType}
        data-uid={uid}
      >
        @{name}
        {isBot && <span class="text-[10px] opacity-60 ml-0.5 font-normal">BOT</span>}
      </span>
    );
    lastIndex = match.index + match[0].length;
  }
  if (lastIndex < props.content.length) {
    parts.push(
      <>{props.content.slice(lastIndex)}</>
    );
  }
  return parts.length > 0 ? <>{parts}</> : <>{props.content}</>;
}

function MarkdownText(props: { content: string; isMine: boolean }) {
  const html = marked.parse(props.content, { breaks: true }) as string;
  return (
    <div
      class={`markdown-body ${props.isMine ? "markdown-mine" : ""}`}
      innerHTML={html}
    />
  );
}

/**
 * Shows an uploading message bubble with a progress bar.
 * For images, displays a local preview via object URL.
 */
function PendingUploadBubble(props: {
  upload: PendingUpload;
  onRetry: (upload: PendingUpload) => void;
  onDismiss: (uploadId: string) => void;
}) {
  const u = props.upload;
  const pct = () => `${u.progress}%`;
  const isError = () => u.status === "error";

  return (
    <div class="flex justify-end px-5">
      <div
        class="max-w-[75%] rounded-2xl rounded-br-md overflow-hidden border"
        classList={{
          "bg-primary text-white border-primary-dark": !isError(),
          "bg-danger/20 border-danger/40": isError(),
        }}
      >
        {u.fileType === "image" && u.previewUrl ? (
          <div class="relative">
            <img
              src={u.previewUrl}
              alt={u.fileName}
              class="max-w-[280px] max-h-[200px] object-cover"
            />
            {/* Progress overlay */}
            <div class="absolute inset-0 bg-black/30 flex flex-col items-center justify-center gap-1">
              {isError() ? (
                <>
                  <span class="text-sm px-2 text-center text-red-200">
                    {u.errorMsg || "上传失败"}
                  </span>
                </>
              ) : (
                <>
                  <Loader2 size={28} class="animate-spin text-white/90" />
                  <span class="text-xs text-white/80">{pct()}</span>
                </>
              )}
            </div>
          </div>
        ) : (
          <div class="flex items-center gap-3 px-4 py-3">
            {isError() ? (
              <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-danger/20 flex items-center justify-center">
                <FileText size={20} class="text-danger" />
              </div>
            ) : (
              <div class="flex-shrink-0">
                <Loader2 size={22} class="animate-spin" />
              </div>
            )}
            <div class="min-w-0 flex-1">
              <p class={`text-sm font-medium truncate ${isError() ? "text-danger" : ""}`}>
                {u.fileName}
              </p>
              <p class={`text-xs ${isError() ? "text-danger/70" : "text-white/70"}`}>
                {isError() ? u.errorMsg || "上传失败" : formatFileSize(u.fileSize)}
              </p>
            </div>
          </div>
        )}

        {/* Progress bar (only when uploading) */}
        <Show when={!isError()}>
          <div class="h-1 bg-black/20">
            <div
              class="h-full bg-white/50 transition-all duration-300"
              style={{ width: pct() }}
            />
          </div>
        </Show>

        {/* Retry / Dismiss buttons on error */}
        <Show when={isError()}>
          <div class="flex border-t border-danger/30">
            <button
              onClick={() => props.onRetry(u)}
              class="flex-1 py-1.5 text-xs font-medium text-danger hover:bg-danger/10 transition-colors flex items-center justify-center gap-1"
            >
              重试
            </button>
            <button
              onClick={() => props.onDismiss(u.id)}
              class="flex-1 py-1.5 text-xs text-text-muted hover:bg-surface-hover transition-colors border-l border-danger/30"
            >
              删除
            </button>
          </div>
        </Show>
      </div>
    </div>
  );
}

function MessageBubble(props: {
  msg: MessageItem;
  isMine: boolean;
  onRecall: () => void;
  onReply: () => void;
  onEdit: () => void;
  onTranslate: () => void;
  isTranslating: boolean;
  translateText?: string;
  onImageClick: (url: string) => void;
}) {
  const [showMenu, setShowMenu] = createSignal(false);
  const [showReadList, setShowReadList] = createSignal(false);
  const [showEditHistory, setShowEditHistory] = createSignal(false);

  interface EditHistoryEntry {
    old_content: string;
    edited_at: number;
  }

  const editHistory = (): EditHistoryEntry[] => {
    try {
      if (props.msg.extra) {
        const extra = JSON.parse(props.msg.extra);
        if (extra.edit_history && Array.isArray(extra.edit_history)) {
          return extra.edit_history;
        }
      }
    } catch {}
    return [];
  };

  const hasEditHistory = () => editHistory().length > 0;

  const activeConv = () => {
    const id = chatStore.activeConvId();
    return chatStore.conversations().find((c) => c.conv_id === id);
  };

  const isGroup = () => activeConv()?.type === "GROUP";

  const canRecall = () => props.isMine && !props.msg.recalled && Date.now() - props.msg.created_at < 2 * 60 * 1000;

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
    return formatMessageSnippet(q.content, q.content_type, q.type);
  };

  const quoteSender = () => {
    const q = quotedMsg();
    if (!q) return "";
    return q.sender_name || "";
  };

  return (
    <>
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
                <MentionText content={props.msg.content} isMine={props.isMine} />
              ) : props.msg.content_type === "markdown" ? (
                <MarkdownText content={props.msg.content} isMine={props.isMine} />
              ) : props.msg.content_type?.startsWith("image/") ? (
                <div class="relative group/img">
                  <img
                    src={(props.msg.content.startsWith("http://") || props.msg.content.startsWith("https://"))
                      ? props.msg.content
                      : getServerApiBase() + props.msg.content}
                    alt="图片"
                    class="max-w-[300px] rounded-lg cursor-pointer"
                    onClick={() => props.onImageClick(props.msg.content)}
                  />
                  <button
                    onClick={async (e) => {
                      e.stopPropagation();
                      await downloadFile(props.msg.content, { filename: `image_${Date.now()}` });
                    }}
                    class="absolute top-1 right-1 p-1 rounded bg-black/40 text-white/90 hover:bg-black/60 opacity-0 group-hover/img:opacity-100 transition-opacity"
                    title="下载"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </button>
                </div>
              ) : props.msg.type === "file" ? (
                <FileMessageDisplay content={props.msg.content} />
              ) : props.msg.type === "voice" ? (
                <div class="flex items-center gap-2">
                  <span>🎵</span>
                  <audio src={props.msg.content} controls class="max-w-[250px] h-8" />
                </div>
              ) : (
                <div class="flex items-center gap-2">
                  <span>📎</span>
                  <span class="text-xs">{props.msg.content_type || props.msg.type}</span>
                </div>
              )}
            </div>

            {/* Message actions */}
            <Show when={showMenu()}>
              <div
                class={`absolute top-0 ${
                  props.isMine ? "-left-[8rem]" : "-right-[8rem]"
                } flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity`}
              >
                <button
                  onClick={props.onReply}
                  class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                  title="回复"
                >
                  <Reply size={14} />
                </button>
                {/* Translate button - only for text messages */}
                <Show when={props.msg.content_type === "text" || props.msg.content_type === "markdown"}>
                  <button
                    onClick={props.onTranslate}
                    disabled={props.isTranslating}
                    class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-text disabled:opacity-40"
                    title="翻译"
                  >
                    {props.isTranslating ? (
                      <Loader2 size={14} class="animate-spin" />
                    ) : (
                      <Languages size={14} />
                    )}
                  </button>
                </Show>
                <Show when={props.isMine}>
                  <button
                    onClick={props.onEdit}
                    class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                    title="编辑"
                  >
                    <Edit3 size={14} />
                  </button>
                  <Show when={canRecall()}>
                    <button
                      onClick={props.onRecall}
                      class="p-1 hover:bg-surface rounded transition-colors text-text-muted hover:text-danger"
                      title="撤回"
                    >
                      <X size={14} />
                    </button>
                  </Show>
                </Show>
              </div>
            </Show>
          </div>

          {/* Translation */}
          <Show when={props.translateText}>
            <div class={`mt-1 mb-0.5 ${props.isMine ? "mr-1 text-right" : "ml-1"}`}>
              <p class={`text-xs px-3 py-0.5 rounded-lg inline-block ${
                props.isMine
                  ? "bg-primary/10 text-primary/80"
                  : "bg-surface text-text-muted"
              }`}>
                AI翻译：{props.translateText}
              </p>
            </div>
          </Show>

          <div
            class={`flex items-center gap-1 mt-0.5 ${
              props.isMine ? "justify-end mr-1" : "ml-1"
            }`}
          >
            <span class="text-xs text-text-muted">
              {formatTime(props.msg.created_at)}
            </span>
            <Show when={props.msg.edited}>
              {hasEditHistory() ? (
                <button
                  onClick={() => setShowEditHistory(true)}
                  class="text-xs text-primary hover:text-primary-dark cursor-pointer transition-colors underline underline-offset-2"
                  title="查看编辑历史"
                >
                  已编辑
                </button>
              ) : (
                <span class="text-xs text-text-muted">已编辑</span>
              )}
            </Show>
            <Show when={props.isMine && !props.msg.edited}>
              <CheckCheck size={12} class="text-text-muted" />
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

      {/* Edit History Modal */}
      <Show when={showEditHistory()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowEditHistory(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[70vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-base font-semibold text-text mb-1">编辑历史</h3>
            <p class="text-xs text-text-muted mb-4">共 {editHistory().length} 次编辑</p>
            <div class="flex-1 overflow-y-auto space-y-3">
              <div class="bg-bg rounded-xl p-3 border border-border">
                <p class="text-xs text-text-muted mb-1">当前内容</p>
                <p class="text-sm text-text whitespace-pre-wrap break-words">{props.msg.content}</p>
              </div>
              <For each={[...editHistory()].reverse()}>
                {(entry, idx) => {
                  const total = editHistory().length;
                  const isOriginal = idx() === total - 1;
                  const editNum = total - idx() - 1;
                  return (
                    <div class="bg-surface-hover rounded-xl p-3 border border-border/50">
                      <p class="text-xs text-text-muted mb-1">
                        {isOriginal ? "原始消息" : `第 ${editNum} 次编辑`} · {formatTime(entry.edited_at)}
                      </p>
                      <p class="text-sm text-text-muted whitespace-pre-wrap break-words">{entry.old_content}</p>
                    </div>
                  );
                }}
              </For>
            </div>
            <button
              onClick={() => setShowEditHistory(false)}
              class="mt-4 w-full px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors"
            >
              关闭
            </button>
          </div>
        </div>
      </Show>
    </>
  );
}

function FileMessageDisplay(props: { content: string }) {
  let fileInfo: { name?: string; size?: number; url?: string } = {};
  try {
    fileInfo = JSON.parse(props.content);
  } catch {
    return (
      <div class="flex items-center gap-2">
        <span>📎</span>
        <span>{props.content}</span>
      </div>
    );
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  };

  const handleDownload = async (e: Event) => {
    e.preventDefault();
    e.stopPropagation();
    if (!fileInfo.url) return;
    try {
      await downloadFile(fileInfo.url, {
        filename: fileInfo.name,
      });
    } catch (err) {
      console.error("Download error:", err);
    }
  };

  // Resolve relative URL to full URL for the link
  const resolveUrl = (url: string) => {
    if (url.startsWith("http://") || url.startsWith("https://")) return url;
    return getServerApiBase() + url;
  };

  return (
    <div class="flex items-center gap-2">
      <span class="text-lg">📁</span>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <a
            href={resolveUrl(fileInfo.url || "")}
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm underline underline-offset-2 hover:no-underline truncate block"
          >
            {fileInfo.name || "未知文件"}
          </a>
          <button
            onClick={handleDownload}
            class="shrink-0 p-1 rounded hover:bg-black/10 transition-colors"
            title="下载"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </button>
        </div>
        {fileInfo.size && (
          <span class="text-xs opacity-70">{formatSize(fileInfo.size)}</span>
        )}
      </div>
    </div>
  );
}