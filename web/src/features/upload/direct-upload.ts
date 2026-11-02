import { z } from "zod";

const draftSchema = z.object({
  id: z.string().uuid(),
});

const sessionSchema = z.object({
  id: z.string().uuid(),
  upload_url: z.string().url(),
});

export async function createDraft(token: string, title: string): Promise<string> {
  const response = await postJSON("/api/account/videos", token, { title });
  if (!response.ok) throw new Error(await errorText(response));
  return draftSchema.parse(await response.json()).id;
}

export async function createUploadSession(
  token: string,
  videoId: string,
  declaredSize: number,
): Promise<{ id: string; uploadUrl: string }> {
  const response = await postJSON(
    `/api/account/videos/${videoId}/upload-sessions`,
    token,
    { declared_size: declaredSize },
  );
  if (!response.ok) throw new Error(await errorText(response));
  const parsed = sessionSchema.parse(await response.json());
  return { id: parsed.id, uploadUrl: parsed.upload_url };
}

const processingSchema = z.object({
  processing_status: z.enum(["WAITING_UPLOAD", "QUEUED", "PROCESSING", "READY", "FAILED"]),
  error: z.object({ code: z.string(), message: z.string() }).optional(),
});

export type ProcessingView = z.infer<typeof processingSchema>;

export async function readProcessing(token: string, videoId: string): Promise<ProcessingView | null> {
  const response = await fetch(`/api/account/videos/${videoId}/processing`, {
    headers: { authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!response.ok) return null;
  return processingSchema.parse(await response.json());
}

export async function retryProcessing(token: string, videoId: string): Promise<void> {
  const response = await postJSON(
    `/api/account/videos/${videoId}/retry-processing`,
    token,
    {},
  );
  if (!response.ok) throw new Error(await errorText(response));
}

export async function completeUpload(token: string, sessionId: string): Promise<void> {
  const response = await postJSON(
    `/api/account/upload-sessions/${sessionId}/complete`,
    token,
    {},
  );
  if (!response.ok) throw new Error(await errorText(response));
}

export function sendBytes(
  url: string,
  file: File,
  onProgress: (percent: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", url);
    xhr.upload.addEventListener("progress", (event) => {
      if (!event.lengthComputable || event.total < 1) return;
      onProgress(Math.round((event.loaded / event.total) * 100));
    });
    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress(100);
        resolve();
        return;
      }
      reject(new Error("upload failed"));
    });
    xhr.addEventListener("error", () => reject(new Error("upload failed")));
    xhr.addEventListener("abort", () => reject(new Error("upload failed")));
    xhr.send(file);
  });
}

async function postJSON(path: string, token: string, body: unknown): Promise<Response> {
  return fetch(path, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });
}

async function errorText(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: { message?: string } };
    return body.error?.message || "Não foi possível preparar o envio.";
  } catch {
    return "Não foi possível preparar o envio.";
  }
}
