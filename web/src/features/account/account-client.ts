import { apiErrorSchema } from "@/features/identity/session-schema";

export function accountRequest(path: string, token: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers);
  headers.set("authorization", `Bearer ${token}`);
  return fetch(`/api/account/${path}`, { ...init, headers });
}

export async function accountError(response: Response, fallback: string): Promise<string> {
  try {
    return apiErrorSchema.parse(await response.json()).error.message;
  } catch {
    return fallback;
  }
}
