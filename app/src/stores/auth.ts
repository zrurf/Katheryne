import { createSignal, createRoot } from "solid-js";
import { api, setTokens, clearTokens, onAuthFailed } from "../services/api";
import { wsService } from "../services/websocket";
import { getActiveServer } from "../services/config";
import {
  startRegistration,
  finishRegistration,
  startLogin,
  finishLogin,
} from "../services/opaque";

let _authFailedUnsub: (() => void) | null = null;

function createAuthStore() {
  const [uid, setUid] = createSignal<string>(localStorage.getItem("uid") || "");
  const [name, setName] = createSignal(localStorage.getItem("name") || "");
  const [avatar, setAvatar] = createSignal(localStorage.getItem("avatar") || "");
  const [isLoggedIn, setIsLoggedIn] = createSignal(!!localStorage.getItem("access_token"));
  const [need2FA, setNeed2FA] = createSignal(false);
  const [tfaToken, setTfaToken] = createSignal("");

  if (_authFailedUnsub) _authFailedUnsub();
  _authFailedUnsub = onAuthFailed(() => {
    wsService.disconnect();
    setUid("");
    setName("");
    setAvatar("");
    setIsLoggedIn(false);
    localStorage.removeItem("uid");
    localStorage.removeItem("name");
    localStorage.removeItem("avatar");
  });

  async function login(phone: string, password: string) {
    const { clientLoginState, startLoginRequest } = startLogin(password);

    const initResp = await api.auth.loginInit(phone, startLoginRequest);

    const loginResult = finishLogin(clientLoginState, initResp.k, phone, password);
    if (!loginResult) {
      throw new Error("登录失败：密码错误或账号不存在");
    }

    const finalResp = await api.auth.loginFinalize(loginResult.finishLoginRequest, initResp.sid);

    if (finalResp.need_2fa) {
      setNeed2FA(true);
      setTfaToken(finalResp.tfa_token || "");
      return { need2FA: true };
    }

    if (finalResp.access_token && finalResp.refresh_token) {
      setTokens(finalResp.access_token, finalResp.refresh_token);
      setUid(finalResp.uid);
      localStorage.setItem("uid", String(finalResp.uid));
      wsService.connect();
      setIsLoggedIn(true);
      await fetchUserInfo(finalResp.uid);
      return { need2FA: false, uid: finalResp.uid };
    }

    throw new Error("登录失败");
  }

  async function verify2FA(code: string) {
    const resp = await api.auth.tfaVerify(tfaToken(), code);
    setTokens(resp.access_token, resp.refresh_token);
    setUid(resp.uid);
    setNeed2FA(false);
    localStorage.setItem("uid", String(resp.uid));
    wsService.connect();
    setIsLoggedIn(true);
    await fetchUserInfo(resp.uid);
    return resp;
  }

  async function register(phone: string, nameStr: string, password: string) {
    const { clientRegistrationState, registrationRequest } = startRegistration(password);

    const initResp = await api.auth.registerInit(phone, registrationRequest);

    const { registrationRecord } = finishRegistration(
      clientRegistrationState,
      initResp.r,
      phone,
      password
    );

    const finalResp = await api.auth.registerFinalize(phone, nameStr, registrationRecord);
    if (!finalResp.result) {
      throw new Error(finalResp.reason || "注册失败");
    }
    return finalResp;
  }

  async function logout() {
    try {
      await api.auth.logout();
    } catch {
      // ignore
    }
    wsService.disconnect();
    clearTokens();
    setUid("");
    setName("");
    setAvatar("");
    setIsLoggedIn(false);
    localStorage.removeItem("uid");
    localStorage.removeItem("name");
    localStorage.removeItem("avatar");
  }

  async function fetchUserInfo(uidParam?: string) {
    const targetUid = uidParam || uid();
    if (!targetUid) return;
    try {
      const info = await api.auth.getUserInfo(targetUid);
      if (targetUid === uid()) {
        setName(info.name);
        setAvatar(info.avatar);
        localStorage.setItem("name", info.name);
        localStorage.setItem("avatar", info.avatar);
      }
      return info;
    } catch {
      // ignore
    }
  }

  function getCurrentServer() {
    return getActiveServer();
  }

  async function updateProfile(name?: string, avatar?: string) {
    try {
      const resp = await api.auth.updateProfile(name, avatar);
      if (resp.name) {
        setName(resp.name);
        localStorage.setItem("name", resp.name);
      }
      if (resp.avatar) {
        setAvatar(resp.avatar);
        localStorage.setItem("avatar", resp.avatar);
      }
    } catch {
      // ignore
    }
  }

  return {
    uid,
    name,
    avatar,
    isLoggedIn,
    need2FA,
    tfaToken,
    login,
    verify2FA,
    register,
    logout,
    fetchUserInfo,
    updateProfile,
    getCurrentServer,
  };
}

export const authStore = createRoot(createAuthStore);