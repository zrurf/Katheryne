import { For, Show, createSignal, createMemo, createResource } from "solid-js";
import { chatStore } from "../../stores/chat";
import { authStore } from "../../stores/auth";
import { api } from "../../services/api";
import type { ConvBotItem, BotInfo } from "../../services/api";
import type { MessageItem } from "../../services/api";
import { Avatar } from "../ui/avatar";
import { downloadFile } from "../../services/download";
import { formatTime } from "../../lib/utils";
import {
  X,
  Users,
  Megaphone,
  Shield,
  UserX,
  VolumeX,
  Crown,
  UserPlus,
  LogOut,
  Hash,
  Settings,
  Search,
  Check,
  Bot,
  Plus,
  Edit,
  Loader2,
} from "lucide-solid";

export function GroupInfoPanel() {
  const [activeSection, setActiveSection] = createSignal<"info" | "members" | "announcements" | "bots" | "files">("info");
  const [showInvite, setShowInvite] = createSignal(false);
  const [inviteSearch, setInviteSearch] = createSignal("");
  const [selectedInvitees, setSelectedInvitees] = createSignal<Set<string>>(new Set<string>());
  const [showAnnouncement, setShowAnnouncement] = createSignal(false);
  const [announcementContent, setAnnouncementContent] = createSignal("");
  const [showMuteModal, setShowMuteModal] = createSignal(false);
  const [muteTarget, setMuteTarget] = createSignal<{ uid: string; name: string } | null>(null);
  const [muteDuration, setMuteDuration] = createSignal(600);
  const [installedBots, setInstalledBots] = createSignal<ConvBotItem[]>([]);
  const [botsLoading, setBotsLoading] = createSignal(false);
  const [showAddBot, setShowAddBot] = createSignal(false);
  const [myBots, setMyBots] = createSignal<BotInfo[]>([]);
  const [myBotsLoading, setMyBotsLoading] = createSignal(false);
  const [hoveredMember, setHoveredMember] = createSignal<{
    uid: string;
    name: string;
    avatar: string;
    role: string;
    nick: string;
    join_time: number;
  } | null>(null);
  const [hoverPosition, setHoverPosition] = createSignal({ x: 0, y: 0 });
  let hoverTimer: ReturnType<typeof setTimeout> | null = null;
  // Group files state
  const [groupFiles, setGroupFiles] = createSignal<MessageItem[]>([]);
  const [filesLoading, setFilesLoading] = createSignal(false);
  // Group editing state
  const [editingGroupName, setEditingGroupName] = createSignal(false);
  const [groupNameInput, setGroupNameInput] = createSignal("");
  const [editingMyNick, setEditingMyNick] = createSignal(false);
  const [myNickInput, setMyNickInput] = createSignal("");
  const [saving, setSaving] = createSignal(false);
  const [uploadingAvatar, setUploadingAvatar] = createSignal(false);
  let avatarInputRef: HTMLInputElement | undefined;

  const handleAvatarChange = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    const g = info();
    if (!g) return;

    if (!file.type.startsWith("image/")) {
      alert("请选择图片文件");
      input.value = "";
      return;
    }

    setUploadingAvatar(true);
    try {
      const uploadResp = await api.oss.uploadWithProgress(file, () => {});
      const proxyPath = `/api/v1/oss/file?key=${encodeURIComponent(uploadResp.oss_index)}`;
      await api.social.updateGroup(g.group_id, undefined, proxyPath, undefined);
      await chatStore.loadGroupInfo(g.group_id);
    } catch (err) {
      alert("上传群头像失败: " + (err as Error).message);
    } finally {
      setUploadingAvatar(false);
      input.value = "";
    }
  };

  const handleSaveGroupName = async () => {
    const g = info();
    const name = groupNameInput().trim();
    if (!g || !name || saving()) return;
    setSaving(true);
    try {
      await api.social.updateGroup(g.group_id, name, undefined, undefined);
      await chatStore.loadGroupInfo(g.group_id);
      setEditingGroupName(false);
    } catch {
      alert("修改群名失败");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveMyNick = async () => {
    const g = info();
    const nick = myNickInput().trim();
    if (!g || saving()) return;
    setSaving(true);
    try {
      await api.social.updateGroupNick(g.group_id, nick);
      await chatStore.loadGroupMembers(g.group_id);
      setEditingMyNick(false);
    } catch {
      alert("修改群昵称失败");
    } finally {
      setSaving(false);
    }
  };

  const info = () => chatStore.groupInfo();
  const members = () => chatStore.groupMembers();
  const announcements = () => chatStore.announcements();

  const myMember = createMemo(() => members().find((m) => m.uid === authStore.uid()));

  const handleMemberMouseEnter = (e: MouseEvent, member: {
    uid: string;
    name: string;
    avatar: string;
    role: string;
    nick: string;
    join_time: number;
  }) => {
    if (hoverTimer) clearTimeout(hoverTimer);
    const target = e.currentTarget as HTMLElement;
    if (!target) return;
    // Capture rect immediately — by the time setTimeout fires, the element
    // may have been removed from the DOM (e.g. when Solid re-renders the list).
    const rect = target.getBoundingClientRect();
    hoverTimer = setTimeout(() => {
      const cardWidth = 224;
      let left = rect.left - cardWidth - 8;
      if (left < 8) left = 8;
      let top = rect.top;
      if (top + 200 > window.innerHeight) top = window.innerHeight - 210;
      setHoverPosition({ x: left, y: top });
      setHoveredMember(member);
    }, 500);
  };

  const handleMemberMouseLeave = () => {
    if (hoverTimer) clearTimeout(hoverTimer);
    hoverTimer = setTimeout(() => {
      setHoveredMember(null);
    }, 200);
  };

  const isOwner = () => info()?.owner === authStore.uid();
  const isAdmin = () => {
    const member = members().find((m) => m.uid === authStore.uid());
    const role = member?.role?.toUpperCase();
    return role === "ADMIN" || role === "OWNER";
  };

  const friends = () => chatStore.friends();
  const memberUids = createMemo(() => new Set(members().map((m) => m.uid)));
  const nonMemberFriends = createMemo(() =>
    friends().filter((f) => !memberUids().has(f.uid))
  );
  const filteredFriends = createMemo(() => {
    const q = inviteSearch().toLowerCase();
    if (!q) return nonMemberFriends();
    return nonMemberFriends().filter(
      (f) => f.name.toLowerCase().includes(q) || f.remark?.toLowerCase().includes(q)
    );
  });

  const toggleInvitee = (uid: string) => {
    setSelectedInvitees((prev) => {
      const next = new Set(prev);
      if (next.has(uid)) {
        next.delete(uid);
      } else {
        next.add(uid);
      }
      return next;
    });
  };

  const handleInvite = async (e: Event) => {
    e.preventDefault();
    if (!info()) return;
    const uids = Array.from(selectedInvitees());
    if (uids.length === 0) return;
    await chatStore.inviteToGroup(info()!.group_id, uids);
    setShowInvite(false);
    setSelectedInvitees(new Set<string>());
    setInviteSearch("");
  };

  const handleCreateAnnouncement = async (e: Event) => {
    e.preventDefault();
    if (!info() || !announcementContent().trim()) return;
    await chatStore.createAnnouncement(info()!.group_id, announcementContent().trim());
    setShowAnnouncement(false);
    setAnnouncementContent("");
  };

  const handleMute = async (e: Event) => {
    e.preventDefault();
    const target = muteTarget();
    if (!info() || !target) return;
    await chatStore.muteMember(info()!.group_id, target.uid, muteDuration());
    setShowMuteModal(false);
    setMuteTarget(null);
  };

  const loadInstalledBots = async () => {
    const g = info();
    if (!g || !chatStore.activeConvId()) return;
    setBotsLoading(true);
    try {
      const resp = await api.bot.getConvBots(chatStore.activeConvId());
      setInstalledBots(resp.list || []);
    } catch {
      setInstalledBots([]);
    } finally {
      setBotsLoading(false);
    }
  };

  createResource(info, loadInstalledBots);

  const loadMyBots = async () => {
    setMyBotsLoading(true);
    try {
      const [myResp, communityResp] = await Promise.all([
        api.bot.listMyBots(),
        api.bot.listCommunityBots(),
      ]);
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
      setMyBots(combined);
    } catch {
      setMyBots([]);
    } finally {
      setMyBotsLoading(false);
    }
  };

  const handleInstallBot = async (botId: string) => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    try {
      await api.bot.install(botId, convId);
      loadInstalledBots();
    } catch {
      // ignore
    }
  };

  const handleUninstallBot = async (botId: string) => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    try {
      await api.bot.uninstall(botId, convId);
      loadInstalledBots();
    } catch {
      // ignore
    }
  };

  const isBotInstalled = (botId: string) =>
    installedBots().some(b => String(b.bot_id) === String(botId) || String(b.bot_id) === botId);

  const loadGroupFiles = async () => {
    const convId = chatStore.activeConvId();
    if (!convId) return;
    setFilesLoading(true);
    try {
      // Fetch recent messages and filter for file/image types
      const resp = await api.message.list(convId, undefined, 100);
      const files = (resp.list || []).filter(
        (m) => m.type === "file" || (m.content_type && m.content_type.startsWith("image/"))
      );
      setGroupFiles(files);
    } catch {
      setGroupFiles([]);
    } finally {
      setFilesLoading(false);
    }
  };

  createResource(activeSection, (section) => {
    if (section === "files") loadGroupFiles();
  });

  return (
    <div class="w-72 h-full bg-bg-secondary border-l border-border flex flex-col shrink-0">
      <div class="p-4 border-b border-border flex items-center justify-between">
        <h2 class="text-sm font-semibold text-text">群组信息</h2>
        <button
          onClick={() => chatStore.setShowGroupPanel(false)}
          class="p-1 hover:bg-surface rounded-lg transition-colors text-text-muted hover:text-text"
        >
          <X size={16} />
        </button>
      </div>

      <div class="flex border-b border-border">
        <button
          class={`flex-1 py-2 text-xs font-medium transition-colors ${
            activeSection() === "info" ? "text-primary border-b-2 border-primary" : "text-text-muted hover:text-text"
          }`}
          onClick={() => setActiveSection("info")}
        >
          信息
        </button>
        <button
          class={`flex-1 py-2 text-xs font-medium transition-colors ${
            activeSection() === "members" ? "text-primary border-b-2 border-primary" : "text-text-muted hover:text-text"
          }`}
          onClick={() => setActiveSection("members")}
        >
          成员 ({members().length})
        </button>
        <button
          class={`flex-1 py-2 text-xs font-medium transition-colors ${
            activeSection() === "announcements" ? "text-primary border-b-2 border-primary" : "text-text-muted hover:text-text"
          }`}
          onClick={() => setActiveSection("announcements")}
        >
          公告
        </button>
        <button
          class={`flex-1 py-2 text-xs font-medium transition-colors ${
            activeSection() === "bots" ? "text-primary border-b-2 border-primary" : "text-text-muted hover:text-text"
          }`}
          onClick={() => setActiveSection("bots")}
        >
          Bot
        </button>
        <button
          class={`flex-1 py-2 text-xs font-medium transition-colors ${
            activeSection() === "files" ? "text-primary border-b-2 border-primary" : "text-text-muted hover:text-text"
          }`}
          onClick={() => setActiveSection("files")}
        >
          文件
        </button>
      </div>

      <div class="flex-1 overflow-y-auto">
        <Show when={activeSection() === "info"}>
          <div class="p-4 space-y-4">
            <div class="flex flex-col items-center gap-3">
              <div class="relative group/avatar cursor-pointer" onClick={() => isAdmin() && avatarInputRef?.click()} title={isAdmin() ? "点击更换群头像" : undefined}>
                <Avatar name={info()?.name} src={info()?.avatar} size="xl" />
                <Show when={isAdmin()}>
                  <div class="absolute inset-0 flex items-center justify-center rounded-full bg-black/30 opacity-0 group-hover/avatar:opacity-100 transition-opacity">
                    {uploadingAvatar() ? (
                      <Loader2 size={20} class="animate-spin text-white" />
                    ) : (
                      <Edit size={18} class="text-white" />
                    )}
                  </div>
                </Show>
              </div>
              <input
                ref={avatarInputRef}
                type="file"
                accept="image/*"
                class="hidden"
                onChange={handleAvatarChange}
              />
              <div class="text-center">
                <Show
                  when={editingGroupName()}
                  fallback={
                    <div class="flex items-center gap-1.5">
                      <h3 class="text-base font-semibold text-text">{info()?.name}</h3>
                      <Show when={isAdmin()}>
                        <button
                          onClick={() => {
                            setGroupNameInput(info()?.name || "");
                            setEditingGroupName(true);
                          }}
                          class="p-1 text-text-muted hover:text-text transition-colors"
                          title="修改群名"
                        >
                          <Edit size={12} />
                        </button>
                      </Show>
                    </div>
                  }
                >
                  <div class="flex items-center gap-1.5">
                    <input
                      type="text"
                      value={groupNameInput()}
                      onInput={(e) => setGroupNameInput(e.currentTarget.value)}
                      onKeyDown={(e) => e.key === "Enter" && handleSaveGroupName()}
                      class="w-40 px-2 py-0.5 bg-bg border border-border rounded text-sm text-text focus:outline-none focus:border-primary"
                      placeholder="群名称"
                    />
                    <button
                      onClick={handleSaveGroupName}
                      disabled={saving()}
                      class="p-1 text-primary hover:bg-primary/10 rounded transition-colors disabled:opacity-40"
                      title="保存"
                    >
                      <Check size={12} />
                    </button>
                    <button
                      onClick={() => setEditingGroupName(false)}
                      class="p-1 text-text-muted hover:text-text transition-colors"
                      title="取消"
                    >
                      <X size={12} />
                    </button>
                  </div>
                </Show>
                <p class="text-xs text-text-muted mt-1">
                  {info()?.member_count} 名成员
                </p>
              </div>
            </div>

            <div class="bg-surface rounded-xl p-3 space-y-2 border border-border">
              <div class="flex items-center justify-between text-xs">
                <span class="text-text-muted">群号</span>
                <span class="text-text font-mono">{info()?.group_id}</span>
              </div>
              <div class="flex items-center justify-between text-xs">
                <span class="text-text-muted">创建时间</span>
                <span class="text-text">{info()?.created_at ? formatTime(info()!.created_at) : "-"}</span>
              </div>
              <div class="flex items-center justify-between text-xs">
                <span class="text-text-muted">验证方式</span>
                <span class="text-text">
                  {info()?.verify_mode === "open" ? "公开" : info()?.verify_mode === "approval" ? "需审核" : "仅邀请"}
                </span>
              </div>
            </div>

            {/* 我的群昵称 */}
            <div class="bg-surface rounded-xl p-3 border border-border">
              <Show
                when={editingMyNick()}
                fallback={
                  <div class="flex items-center justify-between">
                    <div>
                      <span class="text-xs text-text-muted">我的群昵称</span>
                      <p class="text-sm text-text mt-0.5">{myMember()?.nick || "未设置"}</p>
                    </div>
                    <button
                      onClick={() => {
                        setMyNickInput(myMember()?.nick || "");
                        setEditingMyNick(true);
                      }}
                      class="p-1 text-text-muted hover:text-text transition-colors"
                      title="修改群昵称"
                    >
                      <Edit size={14} />
                    </button>
                  </div>
                }
              >
                <div>
                  <span class="text-xs text-text-muted">我的群昵称</span>
                  <div class="flex items-center gap-1.5 mt-1">
                    <input
                      type="text"
                      value={myNickInput()}
                      onInput={(e) => setMyNickInput(e.currentTarget.value)}
                      onKeyDown={(e) => e.key === "Enter" && handleSaveMyNick()}
                      class="flex-1 px-2 py-1 bg-bg border border-border rounded text-sm text-text focus:outline-none focus:border-primary"
                      placeholder="输入群昵称"
                    />
                    <button
                      onClick={handleSaveMyNick}
                      disabled={saving()}
                      class="p-1 text-primary hover:bg-primary/10 rounded transition-colors disabled:opacity-40"
                      title="保存"
                    >
                      <Check size={14} />
                    </button>
                    <button
                      onClick={() => setEditingMyNick(false)}
                      class="p-1 text-text-muted hover:text-text transition-colors"
                      title="取消"
                    >
                      <X size={14} />
                    </button>
                  </div>
                </div>
              </Show>
            </div>

            <div class="space-y-1">
              <button
                onClick={() => setShowInvite(true)}
                class="w-full flex items-center gap-2 px-3 py-2 text-sm text-text hover:bg-surface rounded-lg transition-colors"
              >
                <UserPlus size={16} class="text-text-muted" />
                邀请成员
              </button>
              <Show when={isAdmin()}>
                <button
                  onClick={() => setShowAnnouncement(true)}
                  class="w-full flex items-center gap-2 px-3 py-2 text-sm text-text hover:bg-surface rounded-lg transition-colors"
                >
                  <Megaphone size={16} class="text-text-muted" />
                  发布公告
                </button>
              </Show>
              <Show when={!isOwner()}>
                <button
                  onClick={() => info() && chatStore.leaveGroup(info()!.group_id)}
                  class="w-full flex items-center gap-2 px-3 py-2 text-sm text-danger hover:bg-surface rounded-lg transition-colors"
                >
                  <LogOut size={16} />
                  退出群组
                </button>
              </Show>
            </div>
          </div>
        </Show>

        <Show when={activeSection() === "members"}>
          <div class="p-2">
            <For each={members()}>
              {(member) => (
                <div
                  class="flex items-center gap-3 px-2 py-2 rounded-xl hover:bg-surface transition-colors group relative"
                  onMouseEnter={(e) => handleMemberMouseEnter(e, member)}
                  onMouseLeave={handleMemberMouseLeave}
                >
                  <Avatar name={member.name} src={member.avatar} size="sm" />
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-1.5">
                      <p class="text-sm text-text truncate">
                        {member.nick || member.name}
                      </p>
                      <Show when={member.role === "OWNER"}>
                        <Crown size={12} class="text-warning shrink-0" />
                      </Show>
                      <Show when={member.role === "ADMIN"}>
                        <Shield size={12} class="text-info shrink-0" />
                      </Show>
                    </div>
                    <p class="text-xs text-text-muted">
                      {member.mute_until > Date.now() / 1000 ? "已禁言" : member.role === "OWNER" ? "群主" : member.role === "ADMIN" ? "管理员" : "成员"}
                    </p>
                  </div>
                  <Show when={isAdmin() && member.uid !== authStore.uid() && member.role !== "OWNER"}>
                    <div class="hidden group-hover:flex items-center gap-0.5">
                      <button
                        onClick={() => { setMuteTarget({ uid: member.uid, name: member.name }); setShowMuteModal(true); }}
                        class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted hover:text-warning"
                        title="禁言"
                      >
                        <VolumeX size={14} />
                      </button>
                      <Show when={isOwner()}>
                        <button
                          onClick={() => info() && chatStore.kickMember(info()!.group_id, member.uid)}
                          class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted hover:text-danger"
                          title="踢出"
                        >
                          <UserX size={14} />
                        </button>
                        <button
                          onClick={() => info() && chatStore.transferOwner(info()!.group_id, member.uid)}
                          class="p-1 hover:bg-surface-hover rounded transition-colors text-text-muted hover:text-info"
                          title="转让群主"
                        >
                          <Crown size={14} />
                        </button>
                      </Show>
                    </div>
                  </Show>
                </div>
              )}
            </For>
          </div>

          <Show when={hoveredMember()}>
            <div
              class="fixed z-[100] bg-surface rounded-2xl border border-border shadow-xl p-4 w-56"
              style={{
                left: `${hoverPosition().x}px`,
                top: `${hoverPosition().y}px`,
              }}
              onMouseEnter={() => { if (hoverTimer) clearTimeout(hoverTimer); }}
              onMouseLeave={() => setHoveredMember(null)}
            >
              <div class="flex flex-col items-center gap-3">
                <Avatar name={hoveredMember()!.name} src={hoveredMember()!.avatar} size="xl" />
                <div class="text-center">
                  <p class="text-sm font-semibold text-text">{hoveredMember()!.name}</p>
                  <Show when={hoveredMember()!.nick}>
                    <p class="text-xs text-text-muted mt-0.5">
                      群昵称: {hoveredMember()!.nick}
                    </p>
                  </Show>
                </div>
              </div>
              <div class="mt-3 pt-3 border-t border-border space-y-1.5">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-text-muted">UID</span>
                  <span class="text-text font-mono">{hoveredMember()!.uid}</span>
                </div>
                <div class="flex items-center justify-between text-xs">
                  <span class="text-text-muted">群身份</span>
                  <span class="text-text flex items-center gap-1">
                    <Show when={hoveredMember()!.role === "OWNER"}>
                      <Crown size={12} class="text-warning" />
                      群主
                    </Show>
                    <Show when={hoveredMember()!.role === "ADMIN"}>
                      <Shield size={12} class="text-info" />
                      管理员
                    </Show>
                    <Show when={hoveredMember()!.role !== "OWNER" && hoveredMember()!.role !== "ADMIN"}>
                      成员
                    </Show>
                  </span>
                </div>
                <div class="flex items-center justify-between text-xs">
                  <span class="text-text-muted">入群时间</span>
                  <span class="text-text">{formatTime(hoveredMember()!.join_time)}</span>
                </div>
              </div>
            </div>
          </Show>
        </Show>

        <Show when={activeSection() === "announcements"}>
          <div class="p-3 space-y-2">
            <For each={announcements()}>
              {(ann) => (
                <div class="bg-surface rounded-xl p-3 border border-border">
                  <div class="flex items-center gap-2 mb-1.5">
                    <Avatar name={ann.name} size="sm" />
                    <div>
                      <p class="text-xs font-medium text-text">{ann.name}</p>
                      <p class="text-xs text-text-muted">{formatTime(ann.created_at)}</p>
                    </div>
                  </div>
                  <p class="text-sm text-text whitespace-pre-wrap">{ann.content}</p>
                </div>
              )}
            </For>
            <Show when={announcements().length === 0}>
              <div class="text-center py-8 text-text-muted text-sm">
                <Megaphone size={24} class="mx-auto mb-2 text-text-muted/50" />
                <p>暂无公告</p>
              </div>
            </Show>
          </div>
        </Show>

        <Show when={activeSection() === "bots"}>
          <div class="p-3 space-y-3">
            <Show when={isAdmin()}>
              <button
                onClick={() => { setShowAddBot(!showAddBot()); loadMyBots(); }}
                class="w-full flex items-center justify-center gap-1.5 px-3 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-xs font-medium transition-colors"
              >
                <Plus size={14} />
                {showAddBot() ? "收起" : "添加 Bot"}
              </button>
            </Show>

            <Show when={botsLoading()}>
              <div class="text-center py-4 text-text-muted text-xs">加载中...</div>
            </Show>

            <Show when={!botsLoading() && installedBots().length === 0}>
              <div class="text-center py-6 text-text-muted text-sm">
                <Bot size={24} class="mx-auto mb-2 text-text-muted/50" />
                <p>暂未安装 Bot</p>
              </div>
            </Show>

            <For each={installedBots()}>
              {(bot) => (
                <div class="flex items-center gap-3 p-2 rounded-xl bg-surface border border-border">
                  <Avatar name={bot.name} src={bot.avatar} size="sm" />
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-text truncate">{bot.name}</p>
                    <p class="text-xs text-text-muted truncate">{bot.description || "群 Bot"}</p>
                  </div>
                  <Show when={isAdmin()}>
                    <button
                      onClick={() => handleUninstallBot(String(bot.bot_id))}
                      class="p-1 hover:bg-danger/10 rounded-lg transition-colors text-text-muted hover:text-danger"
                      title="卸载"
                    >
                      <X size={14} />
                    </button>
                  </Show>
                </div>
              )}
            </For>

            <Show when={showAddBot() && isAdmin()}>
              <div class="border-t border-border pt-3">
                <span class="text-xs font-medium text-text-muted mb-2 block">你的 Bot</span>
                <Show when={myBotsLoading()}>
                  <div class="text-center py-3 text-text-muted text-xs">加载中...</div>
                </Show>
                <Show when={!myBotsLoading() && myBots().length === 0}>
                  <div class="text-center py-4 text-text-muted text-xs">
                    没有可用的 Bot，请在设置中创建
                  </div>
                </Show>
                <For each={myBots()}>
                  {(bot) => {
                    const installed = isBotInstalled(String(bot.bot_id));
                    return (
                      <div class="flex items-center gap-3 p-2 rounded-xl hover:bg-surface-hover transition-colors">
                        <Avatar name={bot.name} src={bot.avatar} size="sm" />
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium text-text truncate">{bot.name}</p>
                          <p class="text-xs text-text-muted truncate">{bot.description || ""}</p>
                        </div>
                        <Show when={installed} fallback={
                          <button
                            onClick={() => handleInstallBot(String(bot.bot_id))}
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
          </div>
        </Show>

        <Show when={activeSection() === "files"}>
          <div class="p-3 space-y-3">
            <Show when={filesLoading()}>
              <div class="text-center py-6 text-text-muted text-xs">加载中...</div>
            </Show>
            <Show when={!filesLoading() && groupFiles().length === 0}>
              <div class="text-center py-6 text-text-muted text-xs">暂无共享文件</div>
            </Show>
            <For each={groupFiles()}>
              {(msg) => {
                let fileInfo: { name?: string; size?: number; url?: string } = {};
                try { fileInfo = JSON.parse(msg.content); } catch { /* not json */ }
                const displayName = msg.type === "file"
                  ? (fileInfo.name || "未知文件")
                  : "图片消息";
                const isImage = msg.content_type?.startsWith("image/");
                const fileUrl = isImage ? msg.content : (fileInfo.url || msg.content);
                const fileSize = fileInfo.size || 0;

                const formatSize = (bytes: number) => {
                  if (bytes < 1024) return bytes + " B";
                  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
                  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
                };

                const handleDownload = (e: Event) => {
                  e.preventDefault();
                  if (!fileUrl) return;
                  downloadFile(fileUrl, { filename: displayName });
                };

                return (
                  <div class="flex items-center gap-3 p-2 rounded-xl bg-surface border border-border hover:bg-surface-hover transition-colors">
                    <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-lg">
                      {isImage ? "🖼" : "📁"}
                    </div>
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-text truncate">{displayName}</p>
                      <p class="text-xs text-text-muted">
                        {msg.sender_name} · {formatTime(msg.created_at)}
                        {fileSize > 0 ? ` · ${formatSize(fileSize)}` : ""}
                      </p>
                    </div>
                    <button
                      onClick={handleDownload}
                      class="flex-shrink-0 p-2 hover:bg-surface rounded-lg transition-colors text-text-muted hover:text-primary"
                      title="下载"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                    </button>
                  </div>
                );
              }}
            </For>
          </div>
        </Show>
      </div>

      {/* Invite Modal */}
      <Show when={showInvite()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => { setShowInvite(false); setSelectedInvitees(new Set<string>()); setInviteSearch(""); }}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4 max-h-[80vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-3">邀请成员</h2>
            <div class="relative mb-3">
              <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
              <input
                type="text"
                value={inviteSearch()}
                onInput={(e) => setInviteSearch(e.currentTarget.value)}
                placeholder="搜索好友..."
                class="w-full pl-8 pr-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors"
              />
            </div>
            <div class="flex-1 overflow-y-auto min-h-0 space-y-1 mb-3">
              <Show when={filteredFriends().length > 0} fallback={
                <div class="text-center py-6 text-text-muted text-sm">
                  {inviteSearch() ? "未找到匹配的好友" : "没有可邀请的好友"}
                </div>
              }>
                <For each={filteredFriends()}>
                  {(friend) => {
                    const isSelected = () => selectedInvitees().has(friend.uid);
                    return (
                      <div
                        class={`flex items-center gap-3 px-3 py-2 rounded-xl cursor-pointer transition-colors ${
                          isSelected() ? "bg-primary/10 border border-primary/20" : "hover:bg-bg border border-transparent"
                        }`}
                        onClick={() => toggleInvitee(friend.uid)}
                      >
                        <Avatar name={friend.name} src={friend.avatar} size="sm" />
                        <div class="flex-1 min-w-0">
                          <p class="text-sm text-text truncate">{friend.remark || friend.name}</p>
                        </div>
                        <div class={`w-5 h-5 rounded-md border-2 flex items-center justify-center shrink-0 transition-colors ${
                          isSelected() ? "bg-primary border-primary" : "border-border"
                        }`}>
                          <Show when={isSelected()}>
                            <Check size={12} class="text-white" />
                          </Show>
                        </div>
                      </div>
                    );
                  }}
                </For>
              </Show>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-xs text-text-muted">
                已选 {selectedInvitees().size} 人
              </span>
              <div class="flex gap-2">
                <button type="button" onClick={() => { setShowInvite(false); setSelectedInvitees(new Set<string>()); setInviteSearch(""); }} class="px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button
                  onClick={handleInvite}
                  disabled={selectedInvitees().size === 0}
                  class="px-4 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl text-sm font-medium transition-colors"
                >邀请</button>
              </div>
            </div>
          </div>
        </div>
      </Show>

      {/* Announcement Modal */}
      <Show when={showAnnouncement()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowAnnouncement(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">发布公告</h2>
            <form onSubmit={handleCreateAnnouncement} class="space-y-3">
              <textarea
                value={announcementContent()}
                onInput={(e) => setAnnouncementContent(e.currentTarget.value)}
                placeholder="输入公告内容..."
                rows={4}
                class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text placeholder:text-text-muted focus:outline-none focus:border-primary transition-colors resize-none"
                required
              />
              <div class="flex gap-2">
                <button type="button" onClick={() => setShowAnnouncement(false)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">发布</button>
              </div>
            </form>
          </div>
        </div>
      </Show>

      {/* Mute Modal */}
      <Show when={showMuteModal()}>
        <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={() => setShowMuteModal(false)}>
          <div class="bg-surface rounded-2xl p-6 border border-border w-full max-w-sm mx-4" onClick={(e) => e.stopPropagation()}>
            <h2 class="text-lg font-semibold text-text mb-4">禁言 {muteTarget()?.name}</h2>
            <form onSubmit={handleMute} class="space-y-3">
              <select
                value={muteDuration()}
                onChange={(e) => setMuteDuration(parseInt(e.currentTarget.value))}
                class="w-full px-3 py-2 bg-bg border border-border rounded-xl text-sm text-text focus:outline-none focus:border-primary transition-colors"
              >
                <option value={600}>10 分钟</option>
                <option value={3600}>1 小时</option>
                <option value={86400}>1 天</option>
                <option value={604800}>7 天</option>
              </select>
              <div class="flex gap-2">
                <button type="button" onClick={() => setShowMuteModal(false)} class="flex-1 px-4 py-2 bg-surface-hover hover:bg-bg rounded-xl text-sm text-text transition-colors">取消</button>
                <button type="submit" class="flex-1 px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-xl text-sm font-medium transition-colors">确认</button>
              </div>
            </form>
          </div>
        </div>
      </Show>
    </div>
  );
}