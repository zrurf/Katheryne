import { getServerApiBase } from "./config";

function getApiBase() {
  return `${getServerApiBase()}/api/v1`;
}

function getBotApiBase() {
  return `${getServerApiBase()}/api/v1/bot`;
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
    updateProfile: (name?: string, avatar?: string) =>
      request<{ name: string; avatar: string }>("/auth/user/profile", {
        method: "POST",
        body: JSON.stringify({ name, avatar }),
      }),
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
    updateGroupNick: (groupId: string, nick: string) =>
      request<{ result: boolean }>("/social/group/nick", {
        method: "POST",
        body: JSON.stringify({ group_id: groupId, nick }),
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
    // ========== Template CRUD ==========
    listMyTemplates: () =>
      request<ListMyTemplatesResp>("/templates", {}, getBotApiBase()),
    getTemplate: (templateId: number) =>
      request<GetTemplateResp>(`/templates/${templateId}`, {}, getBotApiBase()),
    createTemplate: (data: CreateTemplateReq) =>
      request<CreateTemplateResp>("/templates", {
        method: "POST",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    updateTemplate: (templateId: number, data: Partial<CreateTemplateReq>) =>
      request<void>(`/templates/${templateId}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    deleteTemplate: (templateId: number) =>
      request<void>(`/templates/${templateId}`, {
        method: "DELETE",
      }, getBotApiBase()),

    // ========== Instance CRUD ==========
    listMyInstances: () =>
      request<ListMyInstancesResp>("/instances", {}, getBotApiBase()),
    getInstance: (instanceId: number) =>
      request<GetInstanceResp>(`/instances/${instanceId}`, {}, getBotApiBase()),
    createInstance: (data: CreateInstanceReq) =>
      request<CreateInstanceResp>("/instances", {
        method: "POST",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    updateInstance: (instanceId: number, data: Partial<CreateInstanceReq>) =>
      request<void>(`/instances/${instanceId}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }, getBotApiBase()),
    deleteInstance: (instanceId: number) =>
      request<void>(`/instances/${instanceId}`, {
        method: "DELETE",
      }, getBotApiBase()),

    // ========== Installation ==========
    getConvBots: (convId: string) =>
      request<GetConvBotsResp>(`/installation/convs/${convId}/bots`, {}, getBotApiBase()),
    install: (botId: string, convId: string, opts?: { template_id?: string; model_provider?: string; model_name?: string; api_key?: string; kb_config?: string }) =>
      request<void>(`/installation/convs/${convId}/bots/install`, {
        method: "POST",
        body: JSON.stringify({ bot_id: botId, ...opts }),
      }, getBotApiBase()),
    batchInstall: (botId: string, convIds: string[], opts?: { template_id?: string; model_provider?: string; model_name?: string; api_key?: string; kb_config?: string }) =>
      request<{ success_count: number; failed_convs: string[] }>(`/installation/bots/install`, {
        method: "POST",
        body: JSON.stringify({ bot_id: botId, conv_ids: convIds, ...opts }),
      }, getBotApiBase()),
    uninstall: (botId: string, convId: string) =>
      request<void>(`/installation/convs/${convId}/bots/uninstall`, {
        method: "POST",
        body: JSON.stringify({ bot_id: botId }),
      }, getBotApiBase()),

    // ========== Community (old API, backward compat) ==========
    listCommunityBots: (keyword?: string) => {
      const path = keyword ? `/community/bots?keyword=${encodeURIComponent(keyword)}` : "/community/bots";
      return request<ListCommunityBotsResp>(path, {}, getBotApiBase());
    },

    // AI features via official bot
    summarize: (convId: string) =>
      request<SummarizeTicket>(`/bot/summarize`, {
        method: "POST",
        body: JSON.stringify({ conv_id: convId }),
      }),
    summarizeResult: (ticket: string) =>
      request<SummarizeResultResp>(`/bot/summarize/result?ticket=${encodeURIComponent(ticket)}`),
    translate: (text: string, targetLang?: string, sourceLang?: string) =>
      request<BotTranslateResp>(`/bot/translate`, {
        method: "POST",
        body: JSON.stringify({ text, target_lang: targetLang, source_lang: sourceLang }),
      }),
    suggestReplies: (convId: string) =>
      request<BotSuggestResp>(`/bot/suggest`, {
        method: "POST",
        body: JSON.stringify({ conv_id: convId }),
      }),
    moderate: (text: string) =>
      request<BotModerateResp>(`/bot/moderate`, {
        method: "POST",
        body: JSON.stringify({ text }),
      }),
  },

  rag: {
    // ========== Knowledge Base CRUD ==========
    listKBs: () =>
      request<ListKBsResp>("/rag/kb/list"),
    getKB: (kbId: string) =>
      request<GetKBResp>(`/rag/kb/${kbId}`),
    createKB: (data: CreateKBReq) =>
      request<CreateKBResp>("/rag/kb/create", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    updateKB: (kbId: string, data: Partial<CreateKBReq>) =>
      request<void>(`/rag/kb/${kbId}/update`, {
        method: "POST",
        body: JSON.stringify(data),
      }),
    deleteKB: (kbId: string) =>
      request<void>(`/rag/kb/${kbId}/delete`, {
        method: "POST",
      }),

    // ========== Document Management ==========
    listDocs: (kbId: string) =>
      request<ListDocsResp>(`/rag/kb/${kbId}/doc/list`),
    getDocStatus: (kbId: string, docId: string) =>
      request<GetDocStatusResp>(`/rag/kb/${kbId}/doc/${docId}`),
    uploadDoc: async (kbId: string, file: File): Promise<UploadDocResp> => {
      // Step 1: Upload file to OSS
      const ossResp = await api.oss.upload(file);
      // Step 2: Notify RAG service with the OSS key + file metadata
      return request<UploadDocResp>(`/rag/kb/${kbId}/doc/upload`, {
        method: "POST",
        body: JSON.stringify({
          oss_key: ossResp.oss_index,
          file_name: file.name,
          content_type: file.type || "application/octet-stream",
          file_size: file.size,
        }),
      });
    },
    deleteDoc: (kbId: string, docId: string) =>
      request<void>(`/rag/kb/${kbId}/doc/${docId}/delete`, {
        method: "POST",
      }),
    getDocChunks: (kbId: string, docId: string) =>
      request<GetDocChunksResp>(`/rag/kb/${kbId}/doc/${docId}/chunks`),

    // ========== External Sync ==========
    triggerSync: (kbId: string) =>
      request<TriggerSyncResp>(`/rag/kb/${kbId}/sync/trigger`, {
        method: "POST",
      }),
    getSyncStatus: (kbId: string, syncId: string) =>
      request<GetSyncStatusResp>(`/rag/kb/${kbId}/sync/${syncId}`),
    listSyncs: (kbId: string) =>
      request<ListSyncsResp>(`/rag/kb/${kbId}/sync/list`),

    // ========== Search ==========
    search: (data: SearchKBReq) =>
      request<SearchKBResp>("/rag/search", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    // ========== KB Health ==========
    getKBHealth: (kbId: string) =>
      request<GetKBHealthResp>(`/rag/kb/${kbId}/health`),

    // ========== Authorization ==========
    authorizeKB: (data: AuthorizeKBReq) =>
      request<void>("/rag/auth/authorize-kb", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    grantBotAccess: (data: GrantBotKBAccessReq) =>
      request<void>("/rag/auth/grant-bot", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    revokeBotAccess: (data: RevokeBotKBAccessReq) =>
      request<void>("/rag/auth/revoke-bot", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    listKBAuthorizations: () =>
      request<ListKBAuthsResp>("/rag/auth/list"),
    listBotKBs: () =>
      request<ListBotKBsResp>("/rag/auth/bot-kbs"),
  },

  oss: {
    upload: async (file: File): Promise<UploadResp> => {
      const formData = new FormData();
      formData.append("file", file);
      const headers: Record<string, string> = {};
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }
      const res = await fetch(`${getApiBase()}/oss/upload`, {
        method: "POST",
        headers,
        body: formData,
      });
      if (!res.ok) {
        let errMsg = `HTTP ${res.status}`;
        try {
          const wrapper: ApiResponse<null> = await res.json();
          if (wrapper.msg) errMsg = wrapper.msg;
        } catch {
          // ignore
        }
        throw new Error(errMsg);
      }
      const wrapper: ApiResponse<UploadResp> = await res.json();
      if (wrapper.code !== 0) {
        throw new Error(wrapper.msg || `Error code: ${wrapper.code}`);
      }
      // Resolve relative URL (returned by backend as a path) to a full URL
      if (wrapper.data.url && wrapper.data.url.startsWith("/")) {
        wrapper.data.url = new URL(wrapper.data.url, getApiBase()).href;
      }
      return wrapper.data;
    },

    /** Upload file with progress callback. Returns UploadResp on success. */
    uploadWithProgress: (
      file: File,
      onProgress: (pct: number) => void
    ): Promise<UploadResp> => {
      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open("POST", `${getApiBase()}/oss/upload`);

        if (accessToken) {
          xhr.setRequestHeader("Authorization", `Bearer ${accessToken}`);
        }

        xhr.upload.addEventListener("progress", (e) => {
          if (e.lengthComputable) {
            onProgress(Math.round((e.loaded / e.total) * 100));
          }
        });

        xhr.addEventListener("load", () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              const wrapper: ApiResponse<UploadResp> = JSON.parse(xhr.responseText);
              if (wrapper.code !== 0) {
                reject(new Error(wrapper.msg || `Error code: ${wrapper.code}`));
              } else {
                // Resolve relative URL (returned by backend as a path) to a full URL
                if (wrapper.data.url && wrapper.data.url.startsWith("/")) {
                  wrapper.data.url = new URL(wrapper.data.url, getApiBase()).href;
                }
                resolve(wrapper.data);
              }
            } catch {
              reject(new Error("Invalid response"));
            }
          } else {
            let errMsg = `HTTP ${xhr.status}`;
            try {
              const wrapper: ApiResponse<null> = JSON.parse(xhr.responseText);
              if (wrapper.msg) errMsg = wrapper.msg;
            } catch {
              // ignore
            }
            reject(new Error(errMsg));
          }
        });

        xhr.addEventListener("error", () => reject(new Error("Network error")));
        xhr.addEventListener("abort", () => reject(new Error("Upload aborted")));

        const formData = new FormData();
        formData.append("file", file);
        xhr.send(formData);
      });
    },
    initiateUpload: (fileName: string, contentType?: string, totalSize?: number) =>
      request<InitiateUploadResp>("/oss/upload/initiate", {
        method: "POST",
        body: JSON.stringify({ file_name: fileName, content_type: contentType, total_size: totalSize }),
      }),
    completeUpload: (uploadId: string, parts: PartInfo[]) =>
      request<UploadResp>("/oss/upload/complete", {
        method: "POST",
        body: JSON.stringify({ upload_id: uploadId, parts }),
      }),
    getDownloadUrl: (objectKey: string, expireSecs?: number) =>
      request<GetDownloadURLResp>("/oss/download-url", {
        method: "POST",
        body: JSON.stringify({ object_key: objectKey, expire_secs: expireSecs }),
      }),
    getConfig: () =>
      request<OssConfig>("/oss/config"),
    /** Upload a single chunk of a multipart upload using raw binary body. */
    uploadPart: async (uploadId: string, partNumber: number, data: Blob): Promise<UploadPartResp> => {
      const headers: Record<string, string> = {
        "Content-Type": "application/octet-stream",
      };
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }
      const queryStr = `?upload_id=${encodeURIComponent(uploadId)}&part_number=${partNumber}`;
      const res = await fetch(`${getApiBase()}/oss/upload/part${queryStr}`, {
        method: "POST",
        headers,
        body: data,
      });
      if (!res.ok) {
        let errMsg = `HTTP ${res.status}`;
        try { const wrapper: ApiResponse<null> = await res.json(); if (wrapper.msg) errMsg = wrapper.msg; } catch { /* ignore */ }
        throw new Error(errMsg);
      }
      const wrapper: ApiResponse<UploadPartResp> = await res.json();
      if (wrapper.code !== 0) throw new Error(wrapper.msg || `Error code: ${wrapper.code}`);
      return wrapper.data;
    },
    /** Upload file using multipart chunked upload for large files. */
    chunkedUpload: async (
      file: File,
      onProgress: (pct: number) => void,
    ): Promise<UploadResp> => {
      const config = await api.oss.getConfig();
      const chunkSize = config.chunk_size || 5 * 1024 * 1024;
      const maxParts = config.max_parts || 100;
      const totalParts = Math.ceil(file.size / chunkSize);
      if (totalParts > maxParts) {
        throw new Error(`文件过大，超过最大分片数限制 (${maxParts})`);
      }

      // Initiate multipart upload
      const initResp = await api.oss.initiateUpload(file.name, file.type, file.size);
      const parts: PartInfo[] = [];

      for (let i = 0; i < totalParts; i++) {
        const start = i * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const blob = file.slice(start, end);
        const partResp = await api.oss.uploadPart(initResp.upload_id, i + 1, blob);
        parts.push({ part_number: partResp.part_number, etag: partResp.etag });
        onProgress(Math.round(((i + 1) / totalParts) * 100));
      }

      const completeResp = await api.oss.completeUpload(initResp.upload_id, parts);
      if (completeResp.url && completeResp.url.startsWith("/")) {
        completeResp.url = new URL(completeResp.url, getApiBase()).href;
      }
      return completeResp;
    },
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

export interface UploadResp {
  filename: string;
  size: number;
  url: string;
  oss_index: string;
  index_id: string;
  expires_at: number;
}

export interface PartInfo {
  part_number: number;
  etag: string;
}

export interface UploadPartResp {
  etag: string;
  part_number: number;
}

export interface OssConfig {
  max_file_size: number;
  chunk_size: number;
  max_parts: number;
}

export interface GetDownloadURLResp {
  url: string;
  expires_at: number;
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

// ============ Bot Template Types ============

export interface BotTemplateItem {
  template_id: number;
  name: string;
  avatar?: string;
  description?: string;
  category?: string;
  version?: string;
  system_prompt?: string;
  welcome_message?: string;
  conversation_style?: string;
  tool_definitions?: string;
  kb_structure?: string;
  config_schema?: string;
  supported_models?: string;
  is_official: boolean;
  tags?: string[];
  status?: string;
  created_at?: number;
}

export interface ListMyTemplatesResp {
  list: BotTemplateItem[];
}

export interface GetTemplateResp {
  template: BotTemplateItem;
}

export interface CreateTemplateReq {
  name: string;
  avatar?: string;
  description?: string;
  category?: string;
  system_prompt?: string;
  welcome_message?: string;
  conversation_style?: string;
  tool_definitions?: string;
  kb_structure?: string;
  config_schema?: string;
  supported_models?: string;
  tags?: string[];
}

export interface CreateTemplateResp {
  template_id: number;
}

// ============ Bot Instance Types ============

export interface BotInstanceItem {
  instance_id: number;
  bot_id: number;
  template_id: number;
  name: string;
  avatar?: string;
  is_self_hosted: boolean;
  hosted_by?: number;
  model_provider?: string;
  model_name?: string;
  kb_config?: string;
  status?: string;
  created_at?: number;
}

export interface ListMyInstancesResp {
  list: BotInstanceItem[];
}

export interface GetInstanceResp {
  instance: BotInstanceItem;
}

export interface CreateInstanceReq {
  template_id: number;
  name?: string;
  avatar?: string;
  is_self_hosted?: boolean;
  hosted_by?: number;
  model_provider?: string;
  model_name?: string;
  api_key?: string;
  api_base_url?: string;
  kb_config?: string;
  instance_config?: string;
}

export interface CreateInstanceResp {
  instance_id: number;
  bot_id: number;
}

// ============ Legacy Bot Types (backward compat) ============

export interface ListMyBotsResp {
  list: BotInfo[];
}

export interface ListCommunityBotsResp {
  hosted_bots: CommunityHostedBot[];
  templates: CommunityTemplate[];
}

export interface CommunityHostedBot {
  type: "hosted";
  instance_id: string;
  bot_id: string;
  name: string;
  avatar?: string;
  description?: string;
  hosted_by: string;
  installed_count: number;
  status: string;
  template_id: string;
  category?: string;
  tags?: string[];
  is_official: boolean;
}

export interface CommunityTemplate {
  type: "template";
  instance_id: string;
  bot_id: string;
  name: string;
  avatar?: string;
  description?: string;
  hosted_by: string;
  installed_count: number;
  status: string;
  template_id: string;
  category?: string;
  tags?: string[];
  is_official: boolean;
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

// ============ RAG Types ============

export interface CreateKBReq {
  name: string;
  description?: string;
  source_type?: string;
  source_config?: string;
}

export interface CreateKBResp {
  kb_id: string;
}

export interface ListKBsResp {
  list: KBItem[];
}

export interface GetKBResp {
  kb_id: string;
  name: string;
  description: string;
  source_type: string;
  source_config: string;
  doc_count: number;
  total_size: number;
  status: string;
  created_at: number;
  updated_at: number;
}

export interface KBItem {
  kb_id: string;
  name: string;
  description: string;
  source_type: string;
  doc_count: number;
  total_size: number;
  status: string;
  created_at: number;
}

export interface ListDocsResp {
  list: DocItem[];
}

export interface GetDocStatusResp {
  doc: DocItem;
}

export interface DocItem {
  doc_id: string;
  kb_id: string;
  file_name: string;
  content_type: string;
  file_size: number;
  status: string;
  chunk_count: number;
  error_msg?: string;
  created_at: number;
  updated_at: number;
}

export interface UploadDocResp {
  doc_id: string;
  status: string;
}

export interface GetDocChunksResp {
  list: DocChunkItem[];
}

export interface DocChunkItem {
  chunk_id: string;
  content: string;
  index: number;
  metadata?: string;
}

export interface TriggerSyncResp {
  sync_id: string;
}

export interface GetSyncStatusResp {
  sync_id: string;
  kb_id: string;
  status: string;
  progress: number;
  error_msg?: string;
  started_at: number;
  completed_at?: number;
}

export interface ListSyncsResp {
  list: SyncItem[];
}

export interface SyncItem {
  sync_id: string;
  kb_id: string;
  source_type: string;
  status: string;
  progress: number;
  started_at: number;
  completed_at?: number;
}

export interface SearchKBReq {
  kb_ids: string[];
  query: string;
  top_k?: number;
  vector_weight?: number;
  graph_weight?: number;
  keyword_weight?: number;
}

export interface SearchKBResp {
  items: RecallItem[];
  kb_counts: Record<string, number>;
  latency_ms: number;
}

export interface RecallItem {
  doc_id: string;
  kb_id: string;
  content: string;
  score: number;
  metadata?: string;
}

export interface GetKBHealthResp {
  kb_id: string;
  status: string;
  doc_count: number;
  total_size: number;
  avg_chunk_size: number;
  issues: string[];
}

export interface AuthorizeKBReq {
  kb_id: string;
  client_id: string;
}

export interface GrantBotKBAccessReq {
  kb_id: string;
  bot_id: number;
}

export interface RevokeBotKBAccessReq {
  kb_id: string;
  bot_id: number;
}

export interface ListKBAuthsResp {
  list: KBAuthItem[];
}

export interface KBAuthItem {
  kb_id: string;
  kb_name: string;
  client_id: string;
  granted_at: number;
}

export interface ListBotKBsResp {
  list: BotKBItem[];
}

export interface BotKBItem {
  kb_id: string;
  kb_name: string;
  client_id: string;
  granted_at: number;
}

// ============ AI Feature Types ============

export interface SummarizeTicket {
  ticket: string;
}

export interface SummarizeResultResp {
  status: string; // pending | processing | completed | error
  result?: BotSummarizeResp;
  error?: string;
}

export interface BotSummarizeResp {
  summary: string;
  key_points: string[];
  action_items: string[];
}

export interface BotTranslateResp {
  text: string;
  source_lang: string;
  target_lang: string;
}

export interface BotSuggestResp {
  suggestions: string[];
}

export interface BotModerateResp {
  safe: boolean;
  reason: string;
}