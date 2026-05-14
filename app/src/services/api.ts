import { getServerApiBase } from "./config";

function getApiBase() {
  return `${getServerApiBase()}/api/v1`;
}

function getBotApiBase() {
  return `${getServerApiBase()}/bot`;
}

let accessToken: string | null = localStorage.getItem("access_token");
let refreshToken: string | null = localStorage.getItem("refresh_token");

type AuthFailedHandler = () => void;
let authFailedHandlers: AuthFailedHandler[] = [];

export function onAuthFailed(handler: AuthFailedHandler) {
  authFailedHandlers.push(handler);
  return () => {
    authFailedHandlers = authFailedHandlers.filter((h) => h !== handler);
  };
}

function notifyAuthFailed() {
  clearTokens();
  authFailedHandlers.forEach((h) => h());
}

export function triggerAuthFailed() {
  notifyAuthFailed();
}

export function setTokens(access: string, refresh: string) {
  accessToken = access;
  refreshToken = refresh;
  localStorage.setItem("access_token", access);
  localStorage.setItem("refresh_token", refresh);
}

export function clearTokens() {
  accessToken = null;
  refreshToken = null;
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

export function getAccessToken(): string | null {
  return accessToken;
}

export function getRefreshToken(): string | null {
  return refreshToken;
}

export async function refreshAccessToken(): Promise<boolean> {
  if (!refreshToken) return false;
  try {
    const res = await fetch(`${getApiBase()}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) return false;
    const wrapper: ApiResponse<RefreshTokenResp> = await res.json();
    if (wrapper.code !== 0) return false;
    const data = wrapper.data;
    setTokens(data.access_token, data.refresh_token);
    return true;
  } catch {
    return false;
  }
}

interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  baseUrl?: string
): Promise<T> {
  const url = baseUrl || getApiBase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  let res = await fetch(`${url}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401 && refreshToken) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${accessToken}`;
      res = await fetch(`${url}${path}`, {
        ...options,
        headers,
      });
    } else {
      notifyAuthFailed();
      throw new Error("登录已过期，请重新登录");
    }
  }

  if (!res.ok) {
    let errMsg = `HTTP ${res.status}`;
    try {
      const wrapper: ApiResponse<null> = await res.json();
      if (wrapper.msg) errMsg = wrapper.msg;
    } catch {
      // ignore parse error
    }
    throw new Error(errMsg);
  }

  const wrapper: ApiResponse<T> = await res.json();
  if (wrapper.code !== 0) {
    throw new Error(wrapper.msg || `Error code: ${wrapper.code}`);
  }
  return wrapper.data;
}

export async function checkServerHealth(apiUrl: string): Promise<boolean> {
  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);
    const res = await fetch(`${apiUrl}/api/v1/auth/user_info/1`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
      signal: controller.signal,
    });
    clearTimeout(timeout);
    return true;
  } catch {
    return false;
  }
}

export async function validateToken(): Promise<boolean> {
  const token = accessToken;
  if (!token) return false;

  try {
    const res = await fetch(`${getApiBase()}/conversation/total_unread`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });
    if (!res.ok) {
      if (res.status === 401 && refreshToken) {
        return await refreshAccessToken();
      }
      return false;
    }
    return true;
  } catch {
    return false;
  }
}

export const api = {
  auth: {
    loginInit: (phone: string, ke1: string) =>
      request<LoginInitResp>("/auth/login/init", {
        method: "POST",
        body: JSON.stringify({ phone, k: ke1 }),
      }),
    loginFinalize: (ke3: string, sessionId: string) =>
      request<LoginFinalizeResp>("/auth/login/finalize", {
        method: "POST",
        body: JSON.stringify({ k: ke3, sid: sessionId }),
      }),
    registerInit: (phone: string, registrationRequest: string) =>
      request<RegisterInitResp>("/auth/register/init", {
        method: "POST",
        body: JSON.stringify({ phone, r: registrationRequest }),
      }),
    registerFinalize: (phone: string, name: string, registrationRecord: string) =>
      request<RegisterFinalizeResp>("/auth/register/finalize", {
        method: "POST",
        body: JSON.stringify({ phone, name, r: registrationRecord }),
      }),
    logout: () =>
      request<void>("/auth/logout", { method: "POST" }),
    refresh: (token: string) =>
      request<RefreshTokenResp>("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: token }),
      }),
    getUserInfo: (uid: string) =>
      request<UserInfo>("/auth/user_info/" + uid),
    tfaVerify: (tfaToken: string, code: string) =>
      request<TFAVerifyResp>("/auth/tfa-verify", {
        method: "POST",
        body: JSON.stringify({ tfa_token: tfaToken, code }),
      }),
  },

  conversation: {
    list: () =>
      request<GetConversationsResp>("/conversation/list"),
    get: (convId: string) =>
      request<GetConversationResp>("/conversation/" + convId),
    getOrCreateSingle: (peerUid: string) =>
      request<GetOrCreateSingleConvResp>("/conversation/single", {
        method: "POST",
        body: JSON.stringify({ peer_uid: peerUid }),
      }),
    delete: (convId: string) =>
      request<void>("/conversation/delete", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId }),
      }),
    mute: (convId: string, mute: boolean) =>
      request<void>("/conversation/mute", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId, mute }),
      }),
    pin: (convId: string, pinned: boolean) =>
      request<void>("/conversation/pin", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId, pinned }),
      }),
    clearUnread: (convId: string) =>
      request<void>("/conversation/clear_unread", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId }),
      }),
    totalUnread: () =>
      request<GetTotalUnreadResp>("/conversation/total_unread"),
  },

  message: {
    list: (convId: string, cursor?: string, limit = 30, direction = "before") => {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      params.set("limit", String(limit));
      params.set("direction", direction);
      return request<GetMessagesResp>(`/message/${convId}?${params}`);
    },
    send: (data: SendMessageReq) =>
      request<SendMessageResp>("/message/send", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    search: (params: SearchMessagesReq) => {
      const qs = new URLSearchParams();
      qs.set("keyword", params.keyword);
      if (params.conv_id) qs.set("conv_id", params.conv_id);
      if (params.sender) qs.set("sender", params.sender);
      if (params.start_time) qs.set("start_time", String(params.start_time));
      if (params.end_time) qs.set("end_time", String(params.end_time));
      qs.set("page", String(params.page || 1));
      qs.set("size", String(params.size || 20));
      return request<SearchMessagesResp>(`/message/search?${qs}`);
    },
    readMembers: (convId: string, msgId: string) =>
      request<GetMessageReadMembersResp>(`/message/${convId}/read/${msgId}`),
    edit: (convId: string, msgId: string, content: string) =>
      request<void>("/message/edit", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId, msg_id: msgId, content }),
      }),
    recall: (convId: string, msgId: string) =>
      request<void>("/message/recall", {
        method: "POST",
        body: JSON.stringify({ conv_id: convId, msg_id: msgId }),
      }),
    sync: (deviceId: string, lastSyncMsgId?: string, limit = 100) => {
      const params = new URLSearchParams();
      params.set("device_id", deviceId);
      if (lastSyncMsgId) params.set("last_sync_msg_id", lastSyncMsgId);
      params.set("limit", String(limit));
      return request<SyncOfflineMessagesResp>(`/message/sync?${params}`);
    },
  },

  social: {
    friends: (group?: string) => {
      const params = group ? `?group=${encodeURIComponent(group)}` : "";
      return request<GetFriendsResp>("/social/friends" + params);
    },
    deleteFriend: (uid: string) =>
      request<DeleteFriendResp>("/social/friend/delete", {
        method: "POST",
        body: JSON.stringify({ uid }),
      }),
    updateRemark: (uid: string, remark: string, groupName?: string) =>
      request<UpdateFriendRemarkResp>("/social/friend/remark", {
        method: "POST",
        body: JSON.stringify({ uid, remark, group_name: groupName }),
      }),
    sendFriendRequest: (uid: string, message?: string) =>
      request<SendFriendResponse>("/social/friend/request", {
        method: "POST",
        body: JSON.stringify({ uid, message }),
      }),
    handleFriendRequest: (id: string, action: "accept" | "reject") =>
      request<HandleFriendResponse>(`/social/friend/request/${id}`, {
        method: "POST",
        body: JSON.stringify({ action }),
      }),
    friendRequests: (type: "sent" | "received" = "received", page = 1, size = 20) =>
      request<GetFriendRequestsResp>(
        `/social/friend/requests?type=${type}&page=${page}&size=${size}`
      ),
    blacklist: () =>
      request<GetBlacklistResp>("/social/blacklist"),
    addBlacklist: (uid: string) =>
      request<AddBlacklistResp>("/social/blacklist/add", {
        method: "POST",
        body: JSON.stringify({ uid }),
      }),
    removeBlacklist: (uid: string) =>
      request<RemoveBlacklistResp>("/social/blacklist/remove", {
        method: "POST",
        body: JSON.stringify({ uid }),
      }),

    createGroup: (name: string, memberUids?: string[], avatar?: string, verifyMode?: string) =>
      request<CreateGroupResp>("/social/group/create", {
        method: "POST",
        body: JSON.stringify({ name, avatar, member_uids: memberUids, verify_mode: verifyMode }),
      }),
    getGroupInfo: (groupId: string) =>
      request<GroupInfoResp>("/social/group/" + groupId),
    updateGroup: (groupId: string, name?: string, avatar?: string, verifyMode?: string) =>
      request<UpdateGroupResp>("/social/group/update", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, name, avatar, verify_mode: verifyMode }),
      }),
    getMembers: (groupId: string, role?: string) => {
      const params = role ? `?role=${role}` : "";
      return request<GetGroupMembersResp>(`/social/group/${groupId}/members${params}`);
    },
    invite: (groupId: string, inviteeUids: string[], message?: string) =>
      request<InviteToGroupResp>("/social/group/invite", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, invitee_uids: inviteeUids, message }),
      }),
    handleInvite: (id: string, action: "accept" | "reject") =>
      request<HandleGroupInviteResp>(`/social/group/invite/${id}`, {
        method: "POST",
        body: JSON.stringify({ action }),
      }),
    invites: () =>
      request<GetGroupInvitesResp>("/social/group/invites"),
    joinGroup: (groupId: string, message?: string) =>
      request<JoinGroupResp>("/social/group/join", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, message }),
      }),
    handleJoinRequest: (id: string, action: "accept" | "reject") =>
      request<HandleGroupJoinResp>(`/social/group/join/${id}`, {
        method: "POST",
        body: JSON.stringify({ action }),
      }),
    joinRequests: (groupId: string, status?: string) => {
      const params = status ? `?status=${status}` : "";
      return request<GetGroupJoinRequestsResp>(`/social/group/${groupId}/requests${params}`);
    },
    kickMember: (groupId: string, uid: string) =>
      request<KickMemberResp>("/social/group/kick", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, uid }),
      }),
    leaveGroup: (groupId: string) =>
      request<LeaveGroupResp>("/social/group/leave", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId }),
      }),
    muteMember: (groupId: string, uid: string, duration: number) =>
      request<MuteMemberResp>("/social/group/mute", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, uid, duration }),
      }),
    transferOwner: (groupId: string, newOwner: string) =>
      request<TransferOwnerResp>("/social/group/transfer", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, new_owner: newOwner }),
      }),
    announcements: (groupId: string, page = 1, size = 20) =>
      request<GetAnnouncementsResp>(
        `/social/group/${groupId}/announcements?page=${page}&size=${size}`
      ),
    createAnnouncement: (groupId: string, content: string) =>
      request<CreateAnnouncementResp>("/social/group/announcement", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, content }),
      }),
  },

  bot: {
    getConvBots: (convId: string) =>
      request<GetConvBotsResp>(`/installation/v1/convs/${convId}/bots`, {}, getBotApiBase()),
    install: (botId: string, convId: string) =>
      request<void>("/installation/v1/install", {
        method: "POST",
        body: JSON.stringify({ bot_id: botId, conv_id: convId }),
      }, getBotApiBase()),
    uninstall: (botId: string, convId: string) =>
      request<void>("/installation/v1/uninstall", {
        method: "POST",
        body: JSON.stringify({ bot_id: botId, conv_id: convId }),
      }, getBotApiBase()),
    listMyBots: () =>
      request<ListMyBotsResp>("/developer/v1/bots", {}, getBotApiBase()),
    getBot: (botId: string) =>
      request<BotInfo>(`/developer/v1/bots/${botId}`, {}, getBotApiBase()),
    createBot: (data: CreateBotReq) =>
      request<BotInfo>("/developer/v1/bots", {
        method: "POST",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    updateBot: (data: UpdateBotReq) =>
      request<BotInfo>("/developer/v1/bots", {
        method: "PUT",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    deleteBot: (botId: string) =>
      request<void>("/developer/v1/bots", {
        method: "DELETE",
        body: JSON.stringify({ bot_id: botId }),
      }, getBotApiBase()),
  },

  oss: {
    initiateUpload: (fileName: string, contentType?: string, totalSize?: number) =>
      request<InitiateUploadResp>("/oss/upload/initiate", {
        method: "POST",
        body: JSON.stringify({ file_name: fileName, content_type: contentType, total_size: totalSize }),
      }),
    completeUpload: (uploadId: string, parts: PartInfo[]) =>
      request<void>("/oss/upload/complete", {
        method: "POST",
        body: JSON.stringify({ upload_id: uploadId, parts }),
      }),
  },
};

// ============ Types ============

export interface LoginInitResp {
  k: string;
  sid: string;
}

export interface LoginFinalizeResp {
  uid: string;
  need_2fa: boolean;
  tfa_token?: string;
  access_token?: string;
  refresh_token?: string;
  expires_at?: number;
}

export interface RegisterInitResp {
  spk: string;
  r: string;
  cid?: string;
}

export interface RegisterFinalizeResp {
  result: boolean;
  reason?: string;
}

export interface RefreshTokenResp {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface TFAVerifyResp {
  uid: string;
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface UserInfo {
  uid: string;
  name: string;
  avatar: string;
  phone: string;
  status: string;
}

export interface ConversationItem {
  conv_id: string;
  type: string;
  name: string;
  avatar: string;
  group_id?: string;
  last_msg_id?: string;
  last_msg_snippet?: string;
  last_msg_time?: number;
  last_msg_sender?: string;
  unread_count: number;
  mute: boolean;
  pinned: boolean;
  uid?: string;
  peer_uid?: string;
}

export interface GetConversationsResp {
  list: ConversationItem[];
}

export interface GetConversationResp {
  conv_id: string;
  type: string;
  name: string;
  avatar: string;
  group_id?: string;
  mute: boolean;
  pinned: boolean;
}

export interface GetOrCreateSingleConvResp {
  conv_id: string;
  created: boolean;
}

export interface GetTotalUnreadResp {
  count: number;
}

export interface MessageItem {
  id: string;
  conv_id: string;
  sender: string;
  sender_name: string;
  sender_avatar: string;
  type: string;
  content: string;
  content_type: string;
  quote_msg_id?: string;
  quote_content?: string;
  recalled: boolean;
  edited: boolean;
  extra?: string;
  created_at: number;
  read_by?: string[];
}

export interface GetMessagesResp {
  list: MessageItem[];
  has_more: boolean;
}

export interface SendMessageReq {
  conv_id: string;
  receiver?: string;
  type?: string;
  content: string;
  content_type?: string;
  quote_msg_id?: string;
  extra?: string;
}

export interface SendMessageResp {
  msg_id: string;
  conv_id: string;
  created_at: number;
}

export interface SearchMessagesReq {
  keyword: string;
  conv_id?: string;
  sender?: string;
  start_time?: number;
  end_time?: number;
  page?: number;
  size?: number;
}

export interface SearchMessagesResp {
  list: MessageItem[];
  total: number;
}

export interface GetMessageReadMembersResp {
  list: ReadMemberItem[];
  total: number;
}

export interface ReadMemberItem {
  uid: string;
  name: string;
  avatar: string;
  read_at: number;
}

export interface SyncOfflineMessagesResp {
  list: MessageItem[];
  has_more: boolean;
  new_last_sync_msg_id: string;
}

export interface FriendItem {
  uid: string;
  name: string;
  avatar: string;
  remark: string;
  status: string;
  online: boolean;
  group_name: string;
}

export interface GetFriendsResp {
  list: FriendItem[];
}

export interface DeleteFriendResp {
  result: boolean;
}

export interface UpdateFriendRemarkResp {
  result: boolean;
}

export interface SendFriendResponse {
  result: boolean;
}

export interface HandleFriendResponse {
  result: boolean;
}

export interface FriendRequestItem {
  id: string;
  uid: string;
  name: string;
  avatar: string;
  message: string;
  status: string;
  created_at: number;
}

export interface GetFriendRequestsResp {
  list: FriendRequestItem[];
  total: number;
}

export interface GetBlacklistResp {
  list: FriendItem[];
}

export interface AddBlacklistResp {
  result: boolean;
}

export interface RemoveBlacklistResp {
  result: boolean;
}

export interface CreateGroupResp {
  group_id: string;
  conv_id: string;
}

export interface GroupInfoResp {
  group_id: string;
  name: string;
  avatar: string;
  owner: string;
  member_count: number;
  status: string;
  verify_mode: string;
  created_at: number;
}

export interface UpdateGroupResp {
  result: boolean;
}

export interface GroupMemberItem {
  uid: string;
  name: string;
  avatar: string;
  role: string;
  nick: string;
  join_time: number;
  mute_until: number;
}

export interface GetGroupMembersResp {
  list: GroupMemberItem[];
}

export interface InviteToGroupResp {
  failed_uids: string[];
}

export interface HandleGroupInviteResp {
  result: boolean;
}

export interface GroupInviteItem {
  id: string;
  group_id: string;
  group_name: string;
  group_avatar: string;
  inviter_uid: string;
  inviter_name: string;
  message: string;
  created_at: number;
}

export interface GetGroupInvitesResp {
  list: GroupInviteItem[];
}

export interface JoinGroupResp {
  result: boolean;
}

export interface HandleGroupJoinResp {
  result: boolean;
}

export interface GroupJoinRequestItem {
  id: string;
  uid: string;
  name: string;
  avatar: string;
  message: string;
  status: string;
  created_at: number;
}

export interface GetGroupJoinRequestsResp {
  list: GroupJoinRequestItem[];
}

export interface KickMemberResp {
  result: boolean;
}

export interface LeaveGroupResp {
  result: boolean;
}

export interface MuteMemberResp {
  result: boolean;
}

export interface TransferOwnerResp {
  result: boolean;
}

export interface AnnouncementItem {
  id: string;
  uid: string;
  name: string;
  content: string;
  pinned: boolean;
  created_at: number;
}

export interface GetAnnouncementsResp {
  list: AnnouncementItem[];
  total: number;
}

export interface CreateAnnouncementResp {
  result: boolean;
}

export interface InitiateUploadResp {
  upload_id: string;
  bucket: string;
  object_key: string;
}

export interface PartInfo {
  part_number: number;
  etag: string;
}

export interface GetConvBotsResp {
  list: ConvBotItem[];
}

export interface ConvBotItem {
  bot_id: string;
  name: string;
  avatar: string;
  description: string;
  enabled: boolean;
}

export interface ListMyBotsResp {
  list: BotInfo[];
}

export interface BotInfo {
  bot_id: string;
  name: string;
  avatar: string;
  description: string;
  webhook_url?: string;
  credential?: string;
  created_at: number;
}

export interface CreateBotReq {
  name: string;
  avatar?: string;
  description?: string;
  webhook_url?: string;
}

export interface UpdateBotReq {
  bot_id: string;
  name?: string;
  avatar?: string;
  description?: string;
  webhook_url?: string;
}