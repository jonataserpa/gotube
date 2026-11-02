import {
  apiErrorSchema,
  forgotResultSchema,
  sessionSchema,
  type AccountSession,
} from "@/features/identity/session-schema";

let current: AccountSession | null = null;
let inflight: Promise<AccountSession | null> | null = null;
const listeners = new Set<() => void>();

export function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function getSession(): AccountSession | null {
  return current;
}

function publish(next: AccountSession | null): void {
  current = next;
  for (const listener of listeners) listener();
}

export function restoreSession(): Promise<AccountSession | null> {
  if (inflight) return inflight;
  inflight = readSession("/api/auth/refresh")
    .catch(() => {
      publish(null);
      return null;
    })
    .finally(() => {
      inflight = null;
    });
  return inflight;
}

export async function login(
  email: string,
  password: string,
): Promise<string | null> {
  return sendCredentials("/api/auth/login", { email, password });
}

export async function registerAccount(
  email: string,
  password: string,
): Promise<string | null> {
  const response = await postJSON("/api/auth/register", { email, password });
  if (response.ok) return null;
  return errorMessage(response, "Não foi possível criar a conta.");
}

export async function requestPasswordReset(email: string): Promise<{
  message: string;
  resetToken: string | null;
} | null> {
  const response = await postJSON("/api/auth/password/forgot", { email });
  if (!response.ok) return null;
  const parsed = forgotResultSchema.safeParse(await response.json());
  if (!parsed.success) return null;
  return { message: parsed.data.message, resetToken: parsed.data.reset_token ?? null };
}

export async function resetPassword(
  token: string,
  password: string,
): Promise<string | null> {
  const response = await postJSON("/api/auth/password/reset", { token, password });
  if (response.ok) return null;
  return errorMessage(response, "Não foi possível redefinir a senha.");
}

export async function logout(): Promise<void> {
  await fetch("/api/auth/logout", { method: "POST" });
  publish(null);
}

async function sendCredentials(
  path: string,
  body: { email: string; password: string },
): Promise<string | null> {
  const response = await postJSON(path, body);
  if (!response.ok) {
    publish(null);
    return errorMessage(response, "Não foi possível entrar.");
  }
  return storeSession(response);
}

async function readSession(path: string): Promise<AccountSession | null> {
  const response = await fetch(path, { method: "POST" });
  if (!response.ok) {
    publish(null);
    return null;
  }
  const failure = await storeSession(response);
  return failure ? null : current;
}

async function storeSession(response: Response): Promise<string | null> {
  try {
    const parsed = sessionSchema.parse(await response.json());
    publish({
      accessToken: parsed.access_token,
      email: parsed.user.email,
      userId: parsed.user.id,
    });
    return null;
  } catch {
    publish(null);
    return "A resposta da sessão não é válida.";
  }
}

async function postJSON(path: string, body: unknown): Promise<Response> {
  return fetch(path, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
}

async function errorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const parsed = apiErrorSchema.parse(await response.json());
    return parsed.error.message;
  } catch {
    return fallback;
  }
}
