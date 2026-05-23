import { For, Show, createSignal, createResource, onMount } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { chatStore } from "../../stores/chat";
import { authStore } from "../../stores/auth";
import { appNav } from "../../services/nav";
import { api } from "../../services/api";
import { Avatar } from "../ui/avatar";
import { formatConversationTime, truncate } from "../../lib/utils";
import type { ConversationItem, FriendItem, FriendRequestItem, GroupInviteItem, ConvBotItem, BotInfo } from "../../services/api";
import {
  MessageSquare,
  Users,
  Bot,
  Search,
  Plus,
  Settings,
  LogOut,
  User,
  Hash,
  Pin,
  Bell,
  BellOff,
  Trash2,
  UserPlus,
  UsersRound,
  Check,
  X,
  PanelLeftClose,
  PanelLeftOpen,
  GripVertical,
  Inbox,
  Clock,
} from "lucide-solid";

export function ChatSidebar() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = createSignal("");
  const [activeTab, setActiveTab] = createSignal<"chats" | "contacts" | "bots">("chats");
  const [collapsed, setCollapsed] = createSignal(false);
  const [sidebarWidth, setSidebarWidth] = createSignal(320);
  const [showAddFriend, setShowAddFriend] = createSignal(false);
  const [friendUid, setFriendUid] = createSignal("");
  const [friendMsg, setFriendMsg] = createSignal("");
  const [showCreateGroup, setShowCreateGroup] = createSignal(false);
  const [groupName, setGroupName] = createSignal("");
  const [selectedFriends, setSelectedFriends] = createSignal<Set<string>>(new Set<string>());
  const [creatingGroup, setCreatingGroup] = createSignal(false);
  const [groupError, setGroupError] = createSignal("");
  const [showFriendRequests, setShowFriendRequests] = createSignal(false);
  const [friendRequests, setFriendRequests] = createSignal<FriendRequestItem[]>([]);
  const [friendRequestsLoading, setFriendRequestsLoading] = createSignal(false);
  const [friendRequestCount, setFriendRequestCount] = createSignal(0);
  const [groupInvites, setGroupInvites] = createSignal<GroupInviteItem[]>([]);
  const [groupInviteCount, setGroupInviteCount] = createSignal(0);
  const [inviteTab, setInviteTab] = createSignal<"friends" | "groups">("friends");

  onMount(() => {
    loadFriendRequests();
    loadGroupInvites();
  });

  let resizeStartX = 0;
  let resizeStartWidth = 0;

  const startResize = (e: MouseEvent) => {
    e.preventDefault();
    resizeStartX = e.clientX;
    resizeStartWidth = sidebarWidth();
    document.addEventListener("mousemove", onResize);
    document.addEventListener("mouseup", stopResize);
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  };

  const onResize = (e: MouseEvent) => {
    const delta = e.clientX - resizeStartX;
    const newWidth = Math.max(200, Math.min(500, resizeStartWidth + delta));
    setSidebarWidth(newWidth);
  };

  const stopResize = () => {
    document.removeEventListener("mousemove", onResize);
    document.removeEventListener("mouseup", stopResize);
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  };

  const [friends] = createResource(showCreateGroup, async () => {
    try {
      const resp = await api.social.friends();
      return resp.list || [];
    } catch {
      return [];
    }
  });

  const toggleFriend = (uid: string) => {
    setSelectedFriends((prev) => {
      const next = new Set(prev);
      if (next.has(uid)) {
        next.delete(uid);
      } else {
        next.add(uid);
      }
      return next;
    });
  };

  const filteredConversations = () => {
    const q = searchQuery().toLowerCase();
    if (!q) return chatStore.conversations();
    return chatStore.conversations().filter(
      (c) => c.name.toLowerCase().includes(q)
    );
  };

  const pinnedConversations = () =>
    filteredConversations().filter((c) => c.pinned);
  const unpinnedConversations = () =>
    filteredConversations().filter((c) => !c.pinned);

  const handleAddFriend = async (e: Event) => {
    e.preventDefault();
    const uid = friendUid().trim();
    if (!uid) return;
    try {
      await api.social.sendFriendRequest(uid, friendMsg());
      setShowAddFriend(false);
      setFriendUid("");
      setFriendMsg("");
    } catch {
      // ignore
    }
  };

  const handleCreateGroup = async (e: Event) => {
    e.preventDefault();
    if (!groupName().trim()) return;
    setCreatingGroup(true);
    setGroupError("");
    try {
      const uids = Array.from(selectedFriends());
      await chatStore.createGroup(groupName().trim(), uids.length > 0 ? uids : undefined);
      setShowCreateGroup(false);
      setGroupName("");
      setSelectedFriends(new Set<string>());
    } catch (err) {
      setGroupError(err instanceof Error ? err.message : "创建失败");
    } finally {
      setCreatingGroup(false);
    }
  };

  const loadFriendRequests = async () => {
    setFriendRequestsLoading(true);
    try {
      const resp = await api.social.friendRequests("received");
      setFriendRequests(resp.list || []);
      setFriendRequestCount(resp.total || 0);
    } catch {
      // ignore
    } finally {
      setFriendRequestsLoading(false);
    }
  };

  const handleAcceptFriend = async (id: string) => {
    try {
      await api.social.handleFriendRequest(id, "accept");
      setFriendRequests((prev) => prev.filter((r) => r.id !== id));
      setFriendRequestCount((prev) => prev - 1);
      chatStore.loadFriends();
    } catch {
      // ignore
    }
  };

  const handleRejectFriend = async (id: string) => {
    try {
      await api.social.handleFriendRequest(id, "reject");
      setFriendRequests((prev) => prev.filter((r) => r.id !== id));
      setFriendRequestCount((prev) => prev - 1);
    } catch {
      // ignore
    }
  };

  const loadGroupInvites = async () => {
    try {
      const resp = await api.social.invites();
      setGroupInvites(resp.list || []);
      setGroupInviteCount(resp.list?.length || 0);
    } catch {
      // ignore
    }
  };

  const handleAcceptInvite = async (id: string) => {
    try {
      await api.social.handleInvite(id, "accept");
      setGroupInvites((prev) => prev.filter((r) => r.id !== id));
      setGroupInviteCount((prev) => prev - 1);
      chatStore.loadConversations();
    } catch {
      // ignore
    }
  };

  const handleRejectInvite = async (id: string) => {
    try {
      await api.social.handleInvite(id, "reject");
      setGroupInvites((prev) => prev.filter((r) => r.id !== id));
      setGroupInviteCount((prev) => prev - 1);
    } catch {
      // ignore
    }
  };

  const openFriendRequests = () => {
    setShowFriendRequests(true);
    setInviteTab("friends");
    loadFriendRequests();
    loadGroupInvites();
  };

  const totalRequestCount = () => friendRequestCount() + groupInviteCount();

  return (
    <div
      class="h-full bg-bg-secondary border-r border-border flex flex-col shrink-0 relative transition-[width] duration-200"
      style={{ width: collapsed() ? "64px" : `${sidebarWidth()}px` }}
    >
      <Show when={collapsed()} fallback={
        <>
          {/* Header */}
          <div class="p-4 border-b border-border">
            <div class="flex items-center justify-between mb-3">
              <h1 class="text-lg font-bold text-text">Katheryne</h1>
              <div class="flex items-center gap-1">
                <button
                  class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text"
                  onClick={() => setShowCreateGroup(true)}
                  title="创建群组"
                >
                  <UsersRound size={18} />
                </button>
                <button
                  class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text"
                  onClick={() => setShowAddFriend(true)}
                  title="添加好友"
                >
                  <UserPlus size={18} />
                </button>
                <button
                  class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text relative"
                  onClick={openFriendRequests}
                  title="好友申请"
                >
                  <Inbox size={18} />
                  <Show when={totalRequestCount() > 0}>
                    <span class="absolute -top-0.5 -right-0.5 w-4 h-4 bg-danger text-white text-xs rounded-full flex items-center justify-center leading-none">
                      {totalRequestCount() > 99 ? "99+" : totalRequestCount()}
                    </span>
                  </Show>
                </button>
                <button
                  class="p-2 hover:bg-surface rounded-lg transition-colors text-text-secondary hover:text-text"
                  onClick={() => appNav.goSettings()}
                  title="设置"
                >
                  <Settings size={18} />
                </button>
                <button
                  class="p-2 hover:bg-surface rounded-lg transition-colors text-text-muted hover:text-text"
                  onClick={() => setCollapsed(true)}
                  title="收起侧栏"
                >
                  <PanelLeftClose size={18} />
                </button>
              </div>
            </div>

            {/* Search */}
            <div class="relative">
              <Search
                size={16}
                class="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted"
              />
              <input
                type="text"
                placeholder="搜索会话..."
                value={searchQuery()}
                onInput={(e) => setSearchQuery(e.currentTarget.value)}
                class="w-full pl-9 pr-4 py-2 bg-surface border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
              />
            </div>
          </div>

          {/* Tabs */}
          <div class="flex border-b border-border px-2">
            <button
              class={`flex-1 py-2.5 text-sm font-medium transition-colors relative ${
                activeTab() === "chats"
                  ? "text-primary"
                  : "text-text-muted hover:text-text"
              }`}
              onClick={() => setActiveTab("chats")}
            >
              <span class="flex items-center justify-center gap-1.5">
                <MessageSquare size={16} />
                消息
              </span>
              <Show when={chatStore.totalUnread() > 0 && activeTab() !== "chats"}>
                <span class="absolute -top-0.5 right-3 px-1.5 py-0.5 bg-primary text-white text-xs rounded-full min-w-[18px] text-center leading-none">
                  {chatStore.totalUnread() > 99 ? "99+" : chatStore.totalUnread()}
                </span>
              </Show>
              {activeTab() === "chats" && (
                <div class="absolute bottom-0 left-1/4 right-1/4 h-0.5 bg-primary rounded-full" />
              )}
            </button>
            <button
              class={`flex-1 py-2.5 text-sm font-medium transition-colors relative ${
                activeTab() === "contacts"
                  ? "text-primary"
                  : "text-text-muted hover:text-text"
              }`}
              onClick={() => setActiveTab("contacts")}
            >
              <span class="flex items-center justify-center gap-1.5">
                <Users size={16} />
                联系人
              </span>
              {activeTab() === "contacts" && (
                <div class="absolute bottom-0 left-1/4 right-1/4 h-0.5 bg-primary rounded-full" />
              )}
            </button>
            <button
              class={`flex-1 py-2.5 text-sm font-medium transition-colors relative ${
                activeTab() === "bots"
                  ? "text-primary"
                  : "text-text-muted hover:text-text"
              }`}
              onClick={() => setActiveTab("bots")}
            >
              <span class="flex items-center justify-center gap-1.5">
                <Bot size={16} />
                Bot
              </span>
              {activeTab() === "bots" && (
                <div class="absolute bottom-0 left-1/4 right-1/4 h-0.5 bg-primary rounded-full" />
              )}
            </button>
          </div>

          {/* Conversation List */}
          <Show when={activeTab() === "chats"}>
            <div class="flex-1 overflow-y-auto">
              <Show when={pinnedConversations().length > 0}>
                <div class="px-3 pt-2">
                  <p class="text-xs font-medium text-text-muted px-2 py-1 flex items-center gap-1">
                    <Pin size={10} />
                    置顶
                  </p>
                  <For each={pinnedConversations()}>
                    {(conv) => <ConversationItem conv={conv} />}
                  </For>
                </div>
              </Show>
              <div class="px-3 pt-1">
                <Show when={pinnedConversations().length > 0}>
                  <p class="text-xs font-medium text-text-muted px-2 py-1">全部</p>
                </Show>
                <For each={unpinnedConversations()}>
                  {(conv) => <ConversationItem conv={conv} />}
                </For>
                <Show when={filteredConversations().length === 0}>
                  <div class="text-center py-8 text-text-muted text-sm">
                    {searchQuery() ? "未找到匹配的会话" : "暂无会话"}
                  </div>
                </Show>
              </div>
            </div>
          </Show>

          <Show when={activeTab() === "contacts"}>
            <ContactsTab />
          </Show>

          <Show when={activeTab() === "bots"}>
            <BotsTab />
          </Show>

          {/* User Info */}
          <div class="p-3 border-t border-border">
            <div class="flex items-center gap-3 px-2">
              <Avatar name={authStore.name()} src={authStore.avatar()} size="sm" />
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-text truncate">
                  {authStore.name() || "用户"}
                </p>
                <p class="text-xs text-text-muted">在线</p>
              </div>
              <button
                class="p-1.5 hover:bg-surface rounded-lg transition-colors text-text-muted hover:text-text"
                onClick={() => appNav.goSettings()}
              >
                <Settings size={16} />
              </button>
            </div>
          </div>
        </>
      }>
        {/* Collapsed Mode */}
        <div class="flex flex-col items-center h-full py-3 gap-1">
          <button
            class="p-2 hover:bg-surface rounded-lg transition-colors text-text-muted hover:text-text mb-2"
            onClick={() => setCollapsed(false)}
            title="展开侧栏"
          >
            <PanelLeftOpen size={20} />
          </button>
          <button
            class={`p-2.5 rounded-xl transition-colors ${
              activeTab() === "chats" ? "bg-primary/10 text-primary" : "text-text-muted hover:bg-surface hover:text-text"
            }`}
            onClick={() => setActiveTab("chats")}
            title="消息"
          >
            <MessageSquare size={20} />
          </button>
          <button
            class={`p-2.5 rounded-xl transition-colors ${
              activeTab() === "contacts" ? "bg-primary/10 text-primary" : "text-text-muted hover:bg-surface hover:text-text"
            }`}
            onClick={() => setActiveTab("contacts")}
            title="联系人"
          >
            <Users size={20} />
          </button>
          <button
            class={`p-2.5 rounded-xl transition-colors ${
              activeTab() === "bots" ? "bg-primary/10 text-primary" : "text-text-muted hover:bg-surface hover:text-text"
            }`}
            onClick={() => setActiveTab("bots")}
            title="Bot"
          >
            <Bot size={20} />
          </button>

          <div class="flex-1 overflow-y-auto w-full px-1.5 mt-2">
            <Show when={activeTab() === "chats"}>
              <For each={chatStore.conversations()}>
                {(conv) => (
                  <div
                    class={`flex justify-center py-1.5 rounded-xl cursor-pointer transition-colors ${
                      chatStore.activeConvId() === conv.conv_id ? "bg-primary/10" : "hover:bg-surface"
                    }`}
                    onClick={() => chatStore.selectConversation(conv.conv_id)}
                    title={conv.name}
                  >
                    <div class="relative">
                      <Avatar name={conv.name} src={conv.avatar} size="sm" />
                      <Show when={conv.unread_count > 0}>
                        <span class="absolute -top-1 -right-1 w-3.5 h-3.5 bg-primary rounded-full border-2 border-bg-secondary" />
                      </Show>
                    </div>
                  </div>
                )}
              </For>
            </Show>
            <Show when={activeTab() === "contacts"}>
              <For each={chatStore.friends()}>
                {(friend) => (
                  <div class="flex justify-center py-1.5 rounded-xl hover:bg-surface cursor-pointer transition-colors" title={friend.remark || friend.name}>
                    <Avatar name={friend.name} src={friend.avatar} size="sm" />
                  </div>
                )}
              </For>
            </Show>
          </div>

          <div class="mt-auto pb-2">
            <Avatar name={authStore.name()} src={authStore.avatar()} size="sm" />
          </div>
        </div>
      </Show>

      {/* Resize Handle */}
      <Show when={!collapsed()}>
        <div
          class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-primary/30 transition-colors z-10"
          onMouseDown={startResize}
        >
          <div class="absolute top-1/2 -translate-y-1/2 -right-1.5 w-3 h-8 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
            <GripVertical size={14} class="text-text-muted" />
          </div>
        </div>
      </Show>

      {/* Add Friend Modal */}
      <Show when={showAddFriend()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowAddFriend(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">添加好友</h2>
            <form onSubmit={handleAddFriend} class="space-y-3">
              <input
                type="text"
                value={friendUid()}
                onInput={(e) => setFriendUid(e.currentTarget.value)}
                placeholder="输入好友 UID"
                class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                required
              />
              <input
                type="text"
                value={friendMsg()}
                onInput={(e) => setFriendMsg(e.currentTarget.value)}
                placeholder="验证消息（可选）"
                class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
              />
              <div class="flex gap-2">
                <button
                  type="button"
                  onClick={() => setShowAddFriend(false)}
                  class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors"
                >
                  发送请求
                </button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Friend Requests & Group Invites Modal */}
      <Show when={showFriendRequests()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowFriendRequests(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4 max-h-[70vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold text-text">通知</h2>
              <button
                onClick={() => setShowFriendRequests(false)}
                class="p-1 hover:bg-surface-hover rounded-lg transition-colors text-text-muted hover:text-text"
              >
                <X size={18} />
              </button>
            </div>
            <div class="flex gap-1 mb-3 bg-bg rounded-lg p-1">
              <button
                class={`flex-1 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  inviteTab() === "friends" ? "bg-surface text-text shadow-sm" : "text-text-muted hover:text-text"
                }`}
                onClick={() => setInviteTab("friends")}
              >
                好友申请
                <Show when={friendRequestCount() > 0}>
                  <span class="ml-1 px-1 py-0.5 bg-danger text-white text-xs rounded-full">
                    {friendRequestCount()}
                  </span>
                </Show>
              </button>
              <button
                class={`flex-1 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  inviteTab() === "groups" ? "bg-surface text-text shadow-sm" : "text-text-muted hover:text-text"
                }`}
                onClick={() => setInviteTab("groups")}
              >
                群邀请
                <Show when={groupInviteCount() > 0}>
                  <span class="ml-1 px-1 py-0.5 bg-danger text-white text-xs rounded-full">
                    {groupInviteCount()}
                  </span>
                </Show>
              </button>
            </div>
            <div class="flex-1 overflow-y-auto min-h-0 space-y-2">
              <Show when={inviteTab() === "friends"}>
                <Show when={!friendRequestsLoading()} fallback={
                  <div class="text-center py-8 text-text-muted text-sm">加载中...</div>
                }>
                  <Show when={friendRequests().length > 0} fallback={
                    <div class="text-center py-8 text-text-muted text-sm">
                      <Inbox size={32} class="mx-auto mb-2 text-text-muted/50" />
                      <p>暂无好友申请</p>
                    </div>
                  }>
                    <For each={friendRequests()}>
                      {(req) => (
                        <div class="flex items-center gap-3 p-3 bg-bg rounded-xl border border-border">
                          <Avatar name={req.name} src={req.avatar} size="md" />
                          <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-text truncate">{req.name}</p>
                            <Show when={req.message}>
                              <p class="text-xs text-text-muted truncate mt-0.5">{req.message}</p>
                            </Show>
                            <p class="text-xs text-text-muted mt-0.5 flex items-center gap-1">
                              <Clock size={10} />
                              {formatConversationTime(req.created_at)}
                            </p>
                          </div>
                          <div class="flex items-center gap-1 shrink-0">
                            <button
                              onClick={() => handleAcceptFriend(req.id)}
                              class="p-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg transition-colors"
                              title="同意"
                            >
                              <Check size={14} />
                            </button>
                            <button
                              onClick={() => handleRejectFriend(req.id)}
                              class="p-1.5 bg-surface-hover hover:bg-bg rounded-lg transition-colors text-text-muted hover:text-danger"
                              title="拒绝"
                            >
                              <X size={14} />
                            </button>
                          </div>
                        </div>
                      )}
                    </For>
                  </Show>
                </Show>
              </Show>
              <Show when={inviteTab() === "groups"}>
                <Show when={groupInvites().length > 0} fallback={
                  <div class="text-center py-8 text-text-muted text-sm">
                    <UsersRound size={32} class="mx-auto mb-2 text-text-muted/50" />
                    <p>暂无群邀请</p>
                  </div>
                }>
                  <For each={groupInvites()}>
                    {(invite) => (
                      <div class="flex items-center gap-3 p-3 bg-bg rounded-xl border border-border">
                        <Avatar name={invite.group_name} src={invite.group_avatar} size="md" />
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium text-text truncate">{invite.group_name}</p>
                          <p class="text-xs text-text-muted truncate mt-0.5">
                            {invite.inviter_name} 邀请你加入群组
                          </p>
                          <Show when={invite.message}>
                            <p class="text-xs text-text-muted truncate mt-0.5">{invite.message}</p>
                          </Show>
                          <p class="text-xs text-text-muted mt-0.5 flex items-center gap-1">
                            <Clock size={10} />
                            {formatConversationTime(invite.created_at)}
                          </p>
                        </div>
                        <div class="flex items-center gap-1 shrink-0">
                          <button
                            onClick={() => handleAcceptInvite(invite.id)}
                            class="p-1.5 bg-primary hover:bg-primary-dark text-white rounded-lg transition-colors"
                            title="同意"
                          >
                            <Check size={14} />
                          </button>
                          <button
                            onClick={() => handleRejectInvite(invite.id)}
                            class="p-1.5 bg-surface-hover hover:bg-bg rounded-lg transition-colors text-text-muted hover:text-danger"
                            title="拒绝"
                          >
                            <X size={14} />
                          </button>
                        </div>
                      </div>
                    )}
                  </For>
                </Show>
              </Show>
            </div>
          </div>
        </div>
      </Show>

      {/* Create Group Modal */}
      <Show when={showCreateGroup()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => { setShowCreateGroup(false); setSelectedFriends(new Set<string>()); setGroupError(""); }}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-md mx-4 max-h-[80vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">创建群组</h2>
            <form onSubmit={handleCreateGroup} class="flex flex-col flex-1 min-h-0 space-y-3">
              <input
                type="text"
                value={groupName()}
                onInput={(e) => setGroupName(e.currentTarget.value)}
                placeholder="群组名称"
                class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
                required
              />
              <div class="flex-1 min-h-0">
                <p class="text-xs text-text-muted mb-2">
                  选择成员 <span class="text-primary">{selectedFriends().size}</span> 人已选
                </p>
                <div class="overflow-y-auto max-h-48 border border-border rounded-xl bg-bg">
                  <Show when={friends.loading}>
                    <p class="text-xs text-text-muted text-center py-4">加载中...</p>
                  </Show>
                  <Show when={!friends.loading && friends()?.length === 0}>
                    <p class="text-xs text-text-muted text-center py-4">暂无好友</p>
                  </Show>
                  <For each={friends()}>
                    {(friend: FriendItem) => {
                      const isSelected = () => selectedFriends().has(friend.uid);
                      return (
                        <button
                          type="button"
                          class="w-full flex items-center gap-3 px-3 py-2 hover:bg-surface-hover transition-colors text-left"
                          onClick={() => toggleFriend(friend.uid)}
                        >
                          <div
                            class={`w-5 h-5 rounded border-2 flex items-center justify-center flex-shrink-0 transition-colors ${
                              isSelected()
                                ? "bg-primary border-primary"
                                : "border-border"
                            }`}
                          >
                            <Show when={isSelected()}>
                              <Check size={12} class="text-white" />
                            </Show>
                          </div>
                          <Avatar name={friend.remark || friend.name} size="sm" />
                          <div class="flex-1 min-w-0">
                            <p class="text-sm text-text truncate">
                              {friend.remark || friend.name}
                            </p>
                          </div>
                        </button>
                      );
                    }}
                  </For>
                </div>
              </div>
              <Show when={groupError()}>
                <p class="text-xs text-danger">{groupError()}</p>
              </Show>
              <div class="flex gap-2">
                <button
                  type="button"
                  onClick={() => { setShowCreateGroup(false); setSelectedFriends(new Set<string>()); setGroupError(""); }}
                  class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={creatingGroup()}
                  class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 text-white rounded-xl text-sm font-medium transition-colors"
                >
                  {creatingGroup() ? "创建中..." : "创建"}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Show>

    </div>
  );
}

function ConversationItem(props: { conv: ConversationItem }) {
  const conv = () => props.conv;
  const isActive = () => chatStore.activeConvId() === conv().conv_id;
  const [showMenu, setShowMenu] = createSignal(false);

  return (
    <div
      class={`flex items-center gap-3 px-2 py-2.5 rounded-xl cursor-pointer transition-all duration-150 mb-0.5 group relative ${
        isActive()
          ? "bg-primary/10 border border-primary/20"
          : "hover:bg-surface border border-transparent"
      }`}
      onClick={() => chatStore.selectConversation(conv().conv_id)}
      onContextMenu={(e) => {
        e.preventDefault();
        setShowMenu(!showMenu());
      }}
    >
      <Avatar
        name={conv().name}
        src={conv().avatar}
        size="md"
      />
      <div class="flex-1 min-w-0">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-text truncate">{conv().name}</p>
          <span class="text-xs text-text-muted shrink-0 ml-2">
            {formatConversationTime(conv().last_msg_time || 0)}
          </span>
        </div>
        <div class="flex items-center justify-between mt-0.5">
          <p class="text-xs text-text-muted truncate flex-1">
            {conv().last_msg_snippet || "暂无消息"}
          </p>
          <Show when={conv().unread_count > 0}>
            <Show when={!conv().mute} fallback={
              <span class="ml-2 w-2 h-2 bg-text-muted rounded-full shrink-0" />
            }>
              <span class="ml-2 px-1.5 py-0.5 bg-primary text-white text-xs font-medium rounded-full min-w-[20px] text-center">
                {conv().unread_count > 99 ? "99+" : conv().unread_count}
              </span>
            </Show>
          </Show>
        </div>
      </div>

      {/* Context Menu */}
      <Show when={showMenu()}>
        <div class="absolute right-2 top-full mt-1 bg-surface border border-border rounded-xl shadow-lg z-50 py-1 min-w-[140px]" onClick={(e) => e.stopPropagation()}>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-sm text-text hover:bg-bg transition-colors"
            onClick={() => { chatStore.togglePin(conv().conv_id, !conv().pinned); setShowMenu(false); }}
          >
            <Pin size={14} />
            {conv().pinned ? "取消置顶" : "置顶"}
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-sm text-text hover:bg-bg transition-colors"
            onClick={() => { chatStore.toggleMute(conv().conv_id, !conv().mute); setShowMenu(false); }}
          >
            {conv().mute ? <Bell size={14} /> : <BellOff size={14} />}
            {conv().mute ? "取消免打扰" : "免打扰"}
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-sm text-danger hover:bg-bg transition-colors"
            onClick={() => { chatStore.deleteConversation(conv().conv_id); setShowMenu(false); }}
          >
            <Trash2 size={14} />
            删除会话
          </button>
        </div>
      </Show>
    </div>
  );
}

function ContactsTab() {
  return (
    <div class="flex-1 overflow-y-auto p-3">
      <For each={chatStore.friends()}>
        {(friend) => (
          <div
            class="flex items-center gap-3 px-2 py-2.5 rounded-xl cursor-pointer hover:bg-surface transition-colors mb-0.5"
            onClick={() => chatStore.startChat(friend.uid)}
          >
            <Avatar
              name={friend.remark || friend.name}
              src={friend.avatar}
              online={friend.online}
              size="md"
            />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-text truncate">
                {friend.remark || friend.name}
              </p>
              <p class="text-xs text-text-muted">
                {friend.online ? "在线" : "离线"}
              </p>
            </div>
          </div>
        )}
      </For>
      <Show when={chatStore.friends().length === 0}>
        <div class="text-center py-8 text-text-muted text-sm">暂无联系人</div>
      </Show>
    </div>
  );
}

function BotsTab() {
  const [bots, setBots] = createSignal<ConvBotItem[]>([]);
  const [loading, setLoading] = createSignal(false);
  const [showBotList, setShowBotList] = createSignal(false);
  const [availableBots, setAvailableBots] = createSignal<BotInfo[]>([]);
  const [availableLoading, setAvailableLoading] = createSignal(false);

  const loadBots = async () => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    setLoading(true);
    try {
      const resp = await api.bot.getConvBots(convId);
      setBots(resp.list || []);
    } catch {
      setBots([]);
    } finally {
      setLoading(false);
    }
  };

  const loadAvailableBots = async () => {
    setAvailableLoading(true);
    try {
      const [myResp, communityResp] = await Promise.all([
        api.bot.listMyBots(),
        api.bot.listCommunityBots(),
      ]);
      // Combine user's bots + community bots, deduplicate by bot_id
      const seen = new Set<string>();
      const combined: typeof myResp.list = [];
      for (const bot of myResp.list) {
        if (!seen.has(bot.bot_id)) {
          seen.add(bot.bot_id);
          combined.push(bot);
        }
      }
      for (const bot of communityResp.list) {
        if (!seen.has(bot.bot_id)) {
          seen.add(bot.bot_id);
          combined.push(bot);
        }
      }
      setAvailableBots(combined);
    } catch {
      setAvailableBots([]);
    } finally {
      setAvailableLoading(false);
    }
  };

  createResource(chatStore.activeConvId, loadBots);

  const handleInstall = async (botId: string) => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    try {
      await api.bot.install(botId, convId);
      loadBots();
    } catch {
      // ignore
    }
  };

  const handleUninstall = async (botId: string) => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    try {
      await api.bot.uninstall(botId, convId);
      loadBots();
    } catch {
      // ignore
    }
  };

  const isInstalled = (botId: string) => bots().some(b => String(b.bot_id) === String(botId) || String(b.bot_id) === botId);

  return (
    <div class="flex-1 overflow-y-auto p-3 space-y-3">
      <Show when={!chatStore.activeConvId()}>
        <div class="text-center py-8 text-text-muted text-sm">
          <Bot size={32} class="mx-auto mb-2 text-text-muted/50" />
          <p>选择一个对话以管理 Bot</p>
        </div>
      </Show>
      <Show when={chatStore.activeConvId()}>
        <div class="flex items-center justify-between mb-1">
          <span class="text-xs font-medium text-text-muted">
            已安装 ({bots().length})
          </span>
          <button
            onClick={() => { setShowBotList(!showBotList()); loadAvailableBots(); }}
            class="text-xs text-primary hover:text-primary-dark transition-colors"
          >
            {showBotList() ? "收起" : "添加 Bot"}
          </button>
        </div>

        <Show when={loading()}>
          <div class="text-center py-4 text-text-muted text-xs">加载中...</div>
        </Show>

        <Show when={!loading() && bots().length === 0}>
          <div class="text-center py-6 text-text-muted text-sm">
            <Bot size={24} class="mx-auto mb-2 text-text-muted/50" />
            <p>暂未安装 Bot</p>
            <p class="text-xs mt-1">点击上方"添加 Bot"</p>
          </div>
        </Show>

        <For each={bots()}>
          {(bot) => (
            <div class="flex items-center gap-3 p-2 rounded-xl hover:bg-surface border border-transparent hover:border-border transition-colors">
              <Avatar name={bot.name} src={bot.avatar} size="sm" />
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-text truncate">{bot.name}</p>
                <p class="text-xs text-text-muted truncate">{bot.description || "已安装的 Bot"}</p>
              </div>
              <button
                onClick={() => handleUninstall(String(bot.bot_id))}
                class="p-1.5 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger"
                title="卸载"
              >
                <X size={14} />
              </button>
            </div>
          )}
        </For>

        <Show when={showBotList()}>
          <div class="border-t border-border pt-3 mt-3">
            <span class="text-xs font-medium text-text-muted mb-2 block">可用 Bot</span>
            <Show when={availableLoading()}>
              <div class="text-center py-3 text-text-muted text-xs">加载中...</div>
            </Show>
            <Show when={!availableLoading() && availableBots().length === 0}>
              <div class="text-center py-4 text-text-muted text-xs">
                没有可用的 Bot，请在设置中创建
              </div>
            </Show>
            <For each={availableBots()}>
              {(bot) => {
                const installed = isInstalled(String(bot.bot_id));
                return (
                  <div class="flex items-center gap-3 p-2 rounded-xl hover:bg-surface border border-transparent transition-colors">
                    <Avatar name={bot.name} src={bot.avatar} size="sm" />
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-text truncate">{bot.name}</p>
                      <p class="text-xs text-text-muted truncate">{bot.description || ""}</p>
                    </div>
                    <Show when={installed} fallback={
                      <button
                        onClick={() => handleInstall(String(bot.bot_id))}
                        class="px-2 py-1 bg-primary hover:bg-primary-dark text-white rounded-lg text-xs font-medium transition-colors"
                      >
                        安装
                      </button>
                    }>
                      <span class="text-xs text-text-muted">已安装</span>
                    </Show>
                  </div>
                );
              }}
            </For>
          </div>
        </Show>
      </Show>
    </div>
  );
}