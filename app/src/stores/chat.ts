import { createSignal, createRoot, createEffect } from "solid-js";
import { api } from "../services/api";
import { wsService } from "../services/websocket";
import { formatMessageSnippet } from "../lib/utils";
import type {
  ConversationItem,
  MessageItem,
  FriendItem,
  GroupInfoResp,
  GroupMemberItem,
  AnnouncementItem,
  ReadMemberItem,
} from "../services/api";
import { authStore } from "./auth";

function createChatStore() {
  const [conversations, setConversations] = createSignal<ConversationItem[]>([]);
  const [activeConvId, setActiveConvId] = createSignal<string>("");
  const [messages, setMessages] = createSignal<MessageItem[]>([]);
  const [hasMore, setHasMore] = createSignal(false);
  const [loading, setLoading] = createSignal(false);
  const [friends, setFriends] = createSignal<FriendItem[]>([]);
  const [totalUnread, setTotalUnread] = createSignal(0);
  const [typingUsers, setTypingUsers] = createSignal<Record<string, string>>({});
  const [groupInfo, setGroupInfo] = createSignal<GroupInfoResp | null>(null);
  const [groupMembers, setGroupMembers] = createSignal<GroupMemberItem[]>([]);
  const [announcements, setAnnouncements] = createSignal<AnnouncementItem[]>([]);
  const [showGroupPanel, setShowGroupPanel] = createSignal(false);
  const [readReceipts, setReadReceipts] = createSignal<Record<string, ReadMemberItem[]>>({});
  const [readReceiptsLoading, setReadReceiptsLoading] = createSignal<Record<string, boolean>>({});
  const [activeConvUnreadCount, setActiveConvUnreadCount] = createSignal(0);
  const [firstUnreadMsgId, setFirstUnreadMsgId] = createSignal<string>("");

  let wsUnsubs: (() => void)[] = [];

  createEffect(() => {
    if (authStore.isLoggedIn()) {
      if (!wsService.isConnected()) {
        wsService.connect();
      }
      loadConversations();
      loadFriends();
      loadTotalUnread();
      setupWS();
    }
  });

  function setupWS() {
    wsUnsubs.forEach((u) => u());
    wsUnsubs = [];

    wsUnsubs.push(
      wsService.on("new_message", (msg) => {
        const data = msg.data as {
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
        };

        const newMsg: MessageItem = {
          id: data.msg_id,
          conv_id: data.conv_id,
          sender: data.sender,
          sender_name: data.sender_name || "",
          sender_avatar: data.sender_avatar || "",
          type: data.msg_type,
          content: data.content,
          content_type: data.content_type,
          quote_msg_id: data.quote_msg_id,
          recalled: false,
          edited: false,
          extra: data.extra,
          created_at: data.created_at,
        };

        if (data.conv_id === activeConvId()) {
          setMessages((prev) => {
            const filtered = prev.filter((m) => !m.id.startsWith("temp_"));
            if (filtered.some((m) => m.id === data.msg_id)) {
              return filtered;
            }
            return [...filtered, newMsg];
          });
        }

        setConversations((prev) =>
          prev.map((c) =>
            c.conv_id === data.conv_id
              ? {
                  ...c,
                  last_msg_id: data.msg_id,
                  last_msg_snippet:
                    formatMessageSnippet(
                      data.content,
                      data.content_type,
                      data.msg_type
                    ),
                  last_msg_time: data.created_at,
                  last_msg_sender: data.sender,
                  unread_count:
                    data.conv_id === activeConvId() ? 0 : c.unread_count + 1,
                }
              : c
          )
        );

        loadTotalUnread();
      }),

      wsService.on("send_message_ack", (msg) => {
        const data = msg.data as {
          msg_id: string;
          conv_id: string;
          created_at: number;
          temp_id?: string;
        };
        setMessages((prev) =>
          prev.map((m) => {
            if (data.temp_id && m.id === data.temp_id) {
              return { ...m, id: data.msg_id, created_at: data.created_at };
            }
            if (!data.temp_id && m.id.startsWith("temp_") && m.conv_id === data.conv_id) {
              return { ...m, id: data.msg_id, created_at: data.created_at };
            }
            return m;
          })
        );
        setConversations((prev) =>
          prev.map((c) =>
            c.conv_id === data.conv_id && c.last_msg_id?.startsWith("temp_")
              ? {
                  ...c,
                  last_msg_id: data.msg_id,
                  last_msg_time: data.created_at,
                }
              : c
          )
        );
      }),

      wsService.on("error", (msg) => {
        const data = msg.data as {
          code: number;
          message: string;
        };
        if (data.code === 500 && data.message?.includes("send message failed")) {
          setMessages((prev) => prev.filter((m) => !m.id.startsWith("temp_")));
        }
        console.error("[WS] error:", data.code, data.message);
      }),

      wsService.on("read_receipt", (msg) => {
        const data = msg.data as {
          conv_id: string;
          uid: string;
          last_read_msg_id: string;
          start_msg_id?: string;
          end_msg_id?: string;
        };
        if (data.conv_id === activeConvId()) {
          const endMsgId = data.end_msg_id || data.last_read_msg_id;
          setMessages((prev) =>
            prev.map((m) => {
              const msgIdNum = parseInt(m.id, 10);
              const endNum = parseInt(endMsgId, 10);
              if (!isNaN(msgIdNum) && !isNaN(endNum) && msgIdNum <= endNum && m.sender === authStore.uid()) {
                const existing = m.read_by || [];
                if (!existing.includes(data.uid)) {
                  return { ...m, read_by: [...existing, data.uid] };
                }
              }
              return m;
            })
          );
        }
      }),

      wsService.on("message_recalled", (msg) => {
        const data = msg.data as {
          conv_id: string;
          msg_id: string;
          uid: string;
        };
        setMessages((prev) =>
          prev.map((m) =>
            m.id === data.msg_id
              ? { ...m, recalled: true, content: "消息已撤回" }
              : m
          )
        );
      }),

      wsService.on("message_edited", (msg) => {
        const data = msg.data as {
          conv_id: string;
          msg_id: string;
          content: string;
          uid: string;
        };
        setMessages((prev) =>
          prev.map((m) => {
            if (m.id === data.msg_id) {
              // 构建新的 extra 字段
              let newExtra = m.extra || "";
              try {
                const existing = m.extra ? JSON.parse(m.extra) : {};
                const editHistory = existing.edit_history || [];
                editHistory.push({
                  old_content: m.content,
                  edited_at: Date.now(),
                });
                newExtra = JSON.stringify({ ...existing, edit_history: editHistory });
              } catch {
                newExtra = JSON.stringify({ edit_history: [{ old_content: m.content, edited_at: Date.now() }] });
              }
              return { ...m, content: data.content, edited: true, extra: newExtra };
            }
            return m;
          })
        );
      }),

      wsService.on("typing", (msg) => {
        const data = msg.data as {
          conv_id: string;
          uid: string;
          status: string;
        };
        if (data.conv_id === activeConvId()) {
          if (data.status === "stop_typing") {
            setTypingUsers((prev) => {
              const next = { ...prev };
              delete next[data.uid];
              return next;
            });
          } else {
            setTypingUsers((prev) => ({
              ...prev,
              [data.uid]: data.status,
            }));
            setTimeout(() => {
              setTypingUsers((prev) => {
                const next = { ...prev };
                if (next[data.uid] === "typing") {
                  delete next[data.uid];
                }
                return next;
              });
            }, 5000);
          }
        }
      }),

      wsService.on("online_status", (msg) => {
        const data = msg.data as { uid: string; status: string };
        setFriends((prev) =>
          prev.map((f) =>
            f.uid === data.uid
              ? { ...f, online: data.status === "online", status: data.status }
              : f
          )
        );
      })
    );
  }

  async function loadConversations() {
    try {
      const resp = await api.conversation.list();
      setConversations(resp.list);
    } catch (e) {
      console.error('[chat] loadConversations failed:', e);
    }
  }

  async function loadFriends() {
    try {
      const resp = await api.social.friends();
      setFriends(resp.list);
    } catch (e) {
      console.error('[chat] loadFriends failed:', e);
    }
  }

  async function loadTotalUnread() {
    try {
      const resp = await api.conversation.totalUnread();
      setTotalUnread(resp.count);
    } catch (e) {
      console.error('[chat] loadTotalUnread failed:', e);
    }
  }

  async function loadMessages(convId: string, cursor?: string) {
    setLoading(true);
    try {
      const resp = await api.message.list(convId, cursor);
      if (cursor) {
        setMessages((prev) => [...resp.list, ...prev]);
      } else {
        setMessages(resp.list);
      }
      setHasMore(resp.has_more);
    } catch (e) {
      console.error('[chat] loadMessages failed:', e);
    } finally {
      setLoading(false);
    }
  }

  async function selectConversation(convId: string) {
    setActiveConvId(convId);
    setMessages([]);
    setGroupInfo(null);
    setGroupMembers([]);
    setAnnouncements([]);
    setShowGroupPanel(false);

    const conv = conversations().find((c) => c.conv_id === convId);
    const unreadCount = conv?.unread_count || 0;
    setActiveConvUnreadCount(unreadCount);
    setFirstUnreadMsgId("");

    await loadMessages(convId);

    if (unreadCount > 0) {
      const msgs = messages();
      if (unreadCount <= msgs.length) {
        setFirstUnreadMsgId(String(msgs[msgs.length - unreadCount].id));
      } else {
        // Unread count exceeds loaded page - load older messages to find the first unread
        const oldestMsg = msgs[0];
        await loadMessages(convId, oldestMsg.id);
        const allMsgs = messages();
        const idx = Math.max(0, allMsgs.length - unreadCount);
        if (idx < allMsgs.length) {
          setFirstUnreadMsgId(String(allMsgs[idx].id));
        }
      }
    }

    if (conv?.type === "GROUP" && conv.group_id) {
      await loadGroupInfo(conv.group_id);
      await loadGroupMembers(conv.group_id);
      await loadAnnouncements(conv.group_id);
    }

    try {
      await api.conversation.clearUnread(convId);
      setConversations((prev) =>
        prev.map((c) =>
          c.conv_id === convId ? { ...c, unread_count: 0 } : c
        )
      );
      loadTotalUnread();
    } catch {
      // ignore
    }
  }

  async function startChat(peerUid: string) {
    try {
      const resp = await api.conversation.getOrCreateSingle(peerUid);
      await loadConversations();
      await selectConversation(resp.conv_id);
    } catch {
      // ignore
    }
  }

  async function sendMessage(
    convId: string,
    receiver: string,
    msgType: string,
    content: string,
    contentType: string,
    quoteMsgId?: string,
    extra?: string
  ) {
    const conv = conversations().find((c) => c.conv_id === convId);
    let actualReceiver = receiver;
    if (!actualReceiver && conv && conv.type !== "GROUP") {
      actualReceiver = "";
    }

    const tempId = `temp_${Date.now()}_${Math.random().toString(36).slice(2)}`;
    const optimisticMsg: MessageItem = {
      id: tempId,
      conv_id: convId,
      sender: authStore.uid() || "",
      sender_name: authStore.name() || "",
      sender_avatar: authStore.avatar() || "",
      type: msgType,
      content,
      content_type: contentType,
      quote_msg_id: quoteMsgId,
      recalled: false,
      edited: false,
      extra: extra || "",
      created_at: Date.now(),
    };

    if (convId === activeConvId()) {
      setMessages((prev) => [...prev, optimisticMsg]);
    }

    setConversations((prev) =>
      prev.map((c) =>
        c.conv_id === convId
          ? {
              ...c,
              last_msg_id: tempId,
              last_msg_snippet:
                formatMessageSnippet(content, contentType, msgType),
              last_msg_time: Date.now(),
              last_msg_sender: authStore.uid() || "",
            }
          : c
      )
    );

    if (wsService.isConnected()) {
      wsService.sendMessage({
        conv_id: convId,
        receiver: actualReceiver,
        msg_type: msgType,
        content,
        content_type: contentType,
        quote_msg_id: quoteMsgId,
        extra: extra || undefined,
        temp_id: tempId,
      });
    } else {
      try {
        const resp = await api.message.send({
          conv_id: convId,
          receiver: actualReceiver,
          type: msgType,
          content,
          content_type: contentType,
          quote_msg_id: quoteMsgId,
          extra: extra || undefined,
        });
        setMessages((prev) =>
          prev.map((m) =>
            m.id === tempId
              ? { ...m, id: resp.msg_id, created_at: resp.created_at }
              : m
          )
        );
      } catch (err) {
        console.error("HTTP sendMessage failed:", err);
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
        throw err;
      }
    }
  }

  async function recallMessage(convId: string, msgId: string) {
    setMessages((prev) =>
      prev.map((m) =>
        m.id === msgId
          ? { ...m, recalled: true, content: "消息已撤回" }
          : m
      )
    );
    try {
      await api.message.recall(convId, msgId);
      wsService.sendRecallMessage(convId, msgId);
    } catch {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === msgId
            ? { ...m, recalled: false, content: m.content }
            : m
        )
      );
    }
  }

  async function editMessage(convId: string, msgId: string, content: string) {
    const originalMsg = messages().find((m) => m.id === msgId);
    const originalContent = originalMsg?.content || "";
    const originalExtra = originalMsg?.extra || "";

    // 构建新的 extra 字段，追加编辑历史
    let newExtra = originalExtra;
    try {
      const existing = originalExtra ? JSON.parse(originalExtra) : {};
      const editHistory = existing.edit_history || [];
      editHistory.push({
        old_content: originalContent,
        edited_at: Date.now(),
      });
      newExtra = JSON.stringify({ ...existing, edit_history: editHistory });
    } catch {
      newExtra = JSON.stringify({ edit_history: [{ old_content: originalContent, edited_at: Date.now() }] });
    }

    setMessages((prev) =>
      prev.map((m) =>
        m.id === msgId
          ? { ...m, content, edited: true, extra: newExtra }
          : m
      )
    );
    try {
      await api.message.edit(convId, msgId, content);
      wsService.sendEditMessage(convId, msgId, content);
    } catch {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === msgId
            ? { ...m, content: originalContent, edited: false }
            : m
        )
      );
    }
  }

  async function sendTyping(convId: string, receiver: string, status: string) {
    wsService.sendTyping(convId, receiver, status);
  }

  async function sendReadReceipt(convId: string, lastReadMsgId: string, startMsgId?: string, endMsgId?: string) {
    wsService.sendReadReceipt(convId, lastReadMsgId, startMsgId, endMsgId);
  }

  async function togglePin(convId: string, pinned: boolean) {
    try {
      await api.conversation.pin(convId, pinned);
      setConversations((prev) =>
        prev.map((c) => (c.conv_id === convId ? { ...c, pinned } : c))
      );
    } catch {
      // ignore
    }
  }

  async function toggleMute(convId: string, mute: boolean) {
    try {
      await api.conversation.mute(convId, mute);
      setConversations((prev) =>
        prev.map((c) => (c.conv_id === convId ? { ...c, mute } : c))
      );
    } catch {
      // ignore
    }
  }

  async function deleteConversation(convId: string) {
    try {
      await api.conversation.delete(convId);
      setConversations((prev) => prev.filter((c) => c.conv_id !== convId));
      if (activeConvId() === convId) {
        setActiveConvId("");
        setMessages([]);
      }
    } catch {
      // ignore
    }
  }

  async function searchMessages(
    keyword: string,
    convId?: string,
    startTime?: number,
    endTime?: number
  ) {
    try {
      const resp = await api.message.search({
        keyword,
        conv_id: convId,
        start_time: startTime,
        end_time: endTime,
      });
      return resp;
    } catch {
      return null;
    }
  }

  async function loadMoreMessages() {
    const msgs = messages();
    if (msgs.length === 0 || !hasMore()) return;
    const oldestMsg = msgs[0];
    await loadMessages(activeConvId(), oldestMsg.id);
  }

  async function loadGroupInfo(groupId: string) {
    try {
      const info = await api.social.getGroupInfo(groupId);
      setGroupInfo(info);
      // Sync group avatar/name to conversations list for sidebar and header
      if (info) {
        setConversations((prev) =>
          prev.map((c) =>
            c.group_id === groupId
              ? { ...c, name: info.name ?? c.name, avatar: info.avatar ?? c.avatar }
              : c
          )
        );
      }
    } catch {
      // ignore
    }
  }

  async function loadGroupMembers(groupId: string) {
    try {
      const resp = await api.social.getMembers(groupId);
      setGroupMembers(resp.list);
    } catch {
      // ignore
    }
  }

  async function loadAnnouncements(groupId: string) {
    try {
      const resp = await api.social.announcements(groupId);
      setAnnouncements(resp.list);
    } catch {
      // ignore
    }
  }

  async function createGroup(name: string, memberUids?: string[]) {
    try {
      const resp = await api.social.createGroup(name, memberUids);
      await loadConversations();
      await selectConversation(resp.conv_id);
      return resp;
    } catch {
      throw new Error("创建群组失败");
    }
  }

  async function kickMember(groupId: string, uid: string) {
    try {
      await api.social.kickMember(groupId, uid);
      setGroupMembers((prev) => prev.filter((m) => m.uid !== uid));
      if (groupInfo()) {
        setGroupInfo((prev) => prev ? { ...prev, member_count: prev.member_count - 1 } : null);
      }
    } catch {
      // ignore
    }
  }

  async function muteMember(groupId: string, uid: string, duration: number) {
    try {
      await api.social.muteMember(groupId, uid, duration);
      setGroupMembers((prev) =>
        prev.map((m) =>
          m.uid === uid
            ? { ...m, mute_until: Math.floor(Date.now() / 1000) + duration }
            : m
        )
      );
    } catch {
      // ignore
    }
  }

  async function transferOwner(groupId: string, newOwner: string) {
    try {
      await api.social.transferOwner(groupId, newOwner);
      await loadGroupInfo(groupId);
      await loadGroupMembers(groupId);
    } catch {
      // ignore
    }
  }

  async function createAnnouncement(groupId: string, content: string) {
    try {
      await api.social.createAnnouncement(groupId, content);
      await loadAnnouncements(groupId);
    } catch {
      // ignore
    }
  }

  async function inviteToGroup(groupId: string, inviteeUids: string[]) {
    try {
      await api.social.invite(groupId, inviteeUids);
      await loadGroupMembers(groupId);
    } catch {
      // ignore
    }
  }

  async function leaveGroup(groupId: string) {
    try {
      await api.social.leaveGroup(groupId);
      setConversations((prev) => prev.filter((c) => c.group_id !== groupId));
      if (activeConvId()) {
        const conv = conversations().find((c) => c.group_id === groupId);
        if (conv && activeConvId() === conv.conv_id) {
          setActiveConvId("");
          setMessages([]);
          setGroupInfo(null);
          setGroupMembers([]);
        }
      }
    } catch {
      // ignore
    }
  }

  async function loadReadMembers(convId: string, msgId: string) {
    const key = `${convId}:${msgId}`;
    if (readReceiptsLoading()[key]) return;
    setReadReceiptsLoading((prev) => ({ ...prev, [key]: true }));
    try {
      const resp = await api.message.readMembers(convId, msgId);
      setReadReceipts((prev) => ({ ...prev, [key]: resp.list }));
    } catch {
      // ignore
    } finally {
      setReadReceiptsLoading((prev) => ({ ...prev, [key]: false }));
    }
  }

  return {
    conversations,
    activeConvId,
    messages,
    hasMore,
    loading,
    friends,
    totalUnread,
    typingUsers,
    groupInfo,
    groupMembers,
    announcements,
    showGroupPanel,
    setShowGroupPanel,
    toggleGroupPanel() {
      setShowGroupPanel(!showGroupPanel());
    },
    readReceipts,
    readReceiptsLoading,
    activeConvUnreadCount,
    firstUnreadMsgId,
    loadConversations,
    loadFriends,
    loadTotalUnread,
    loadMessages,
    selectConversation,
    startChat,
    sendMessage,
    recallMessage,
    editMessage,
    sendTyping,
    sendReadReceipt,
    togglePin,
    toggleMute,
    deleteConversation,
    searchMessages,
    loadMoreMessages,
    loadGroupInfo,
    loadGroupMembers,
    loadAnnouncements,
    createGroup,
    kickMember,
    muteMember,
    transferOwner,
    createAnnouncement,
    inviteToGroup,
    leaveGroup,
    loadReadMembers,
  };
}

export const chatStore = createRoot(createChatStore);