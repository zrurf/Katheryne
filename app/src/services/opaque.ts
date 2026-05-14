import * as opaque from "@serenity-kit/opaque";

function padBase64url(s: string): string {
  const rem = s.length % 4;
  if (rem === 2) return s + "==";
  if (rem === 3) return s + "=";
  return s;
}

function unpadBase64url(s: string): string {
  return s.replace(/=+$/, "");
}

export interface RegistrationStartResult {
  clientRegistrationState: string;
  registrationRequest: string;
}

export interface RegistrationFinishResult {
  registrationRecord: string;
  exportKey: string;
}

export interface LoginStartResult {
  clientLoginState: string;
  startLoginRequest: string;
}

export interface LoginFinishResult {
  finishLoginRequest: string;
  sessionKey: string;
  exportKey: string;
}

export function startRegistration(password: string): RegistrationStartResult {
  const { clientRegistrationState, registrationRequest } =
    opaque.client.startRegistration({ password });
  return {
    clientRegistrationState,
    registrationRequest: padBase64url(registrationRequest),
  };
}

export function finishRegistration(
  clientRegistrationState: string,
  registrationResponse: string,
  phone: string,
  password: string
): RegistrationFinishResult {
  const { registrationRecord, exportKey } = opaque.client.finishRegistration({
    clientRegistrationState,
    registrationResponse: unpadBase64url(registrationResponse),
    password,
    identifiers: {
      client: phone,
      server: "katheryne"
    },
    keyStretching: {
      "argon2id-custom": {
        "iterations": 3,
        "memory": 64 * 1024,
        "parallelism": 4
      }
    }
  });
  return {
    registrationRecord: padBase64url(registrationRecord),
    exportKey,
  };
}

export function startLogin(password: string): LoginStartResult {
  const { clientLoginState, startLoginRequest } =
    opaque.client.startLogin({ password });
  return {
    clientLoginState,
    startLoginRequest: padBase64url(startLoginRequest),
  };
}

export function finishLogin(
  clientLoginState: string,
  loginResponse: string,
  phone: string,
  password: string
): LoginFinishResult | null {
  const result = opaque.client.finishLogin({
    clientLoginState,
    loginResponse: unpadBase64url(loginResponse),
    password,
    identifiers: {
      client: phone,
      server: "katheryne"
    },
    keyStretching: {
      "argon2id-custom": {
        "iterations": 3,
        "memory": 64 * 1024,
        "parallelism": 4
      }
    }
  });
  if (!result) return null;
  const { finishLoginRequest, sessionKey, exportKey } = result;
  return {
    finishLoginRequest: padBase64url(finishLoginRequest),
    sessionKey,
    exportKey,
  };
}