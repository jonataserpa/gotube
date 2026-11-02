import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { SiteShell } from "@/components/layout/site-shell";
import { logout } from "@/features/identity/session";

afterEach(async () => {
  await logout().catch(() => undefined);
  cleanup();
  vi.unstubAllGlobals();
});

test("signed-in new video opens the draft flow and keeps search disabled", async () => {
  stubSession();
  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );

  const create = await screen.findByRole("button", { name: "Novo vídeo" });
  expect(create).toHaveProperty("disabled", false);
  fireEvent.click(create);

  expect(await screen.findByRole("textbox", { name: "Título" })).toBeDefined();
  expect(
    screen.getByRole("textbox", { name: "Pesquisar vídeos, canais e tópicos" }),
  ).toHaveProperty("disabled", true);
});

test("partial and finished bytes do not claim the video is processed", async () => {
  let releaseComplete: (response: Response) => void = () => undefined;
  const complete = new Promise<Response>((resolve) => {
    releaseComplete = resolve;
  });
  const calls: string[] = [];
  stubSession((input) => {
    const url = String(input);
    calls.push(url);
    if (url.endsWith("/api/account/videos")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000aa",
        title: "Aula de Go",
        duration_seconds: 0,
        visibility: "PRIVATE",
        publication_status: "DRAFT",
        processing_status: "WAITING_UPLOAD",
      }, { status: 201 });
    }
    if (url.endsWith("/upload-sessions")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000bb",
        upload_url: "https://storage.example/videos/source",
        expires_at: "2026-10-14T20:00:00Z",
        max_bytes: 5,
      }, { status: 201 });
    }
    if (url.endsWith("/complete")) return complete;
    return Response.json({ error: { code: "NOT_FOUND", message: "rota não encontrada" } }, { status: 404 });
  });
  installXHR({ loaded: 2, total: 5, finish: false });

  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );
  fireEvent.click(await screen.findByRole("button", { name: "Novo vídeo" }));
  fireEvent.change(await screen.findByRole("textbox", { name: "Título" }), {
    target: { value: "Aula de Go" },
  });
  fireEvent.change(screen.getByLabelText("Arquivo"), {
    target: { files: [new File(["hello"], "aula.mp4", { type: "video/mp4" })] },
  });
  fireEvent.click(screen.getByRole("button", { name: "Enviar" }));

  const progress = await screen.findByRole("progressbar");
  expect(progress.getAttribute("aria-valuenow")).toBe("40");
  expect(screen.queryByText(/processado|publicado|pronto para assistir/i)).toBeNull();
  expect(calls.some((url) => url.endsWith("/complete"))).toBe(false);

  finishXHR();
  expect(await screen.findByText(/arquivo recebido/i)).toBeDefined();
  expect(screen.queryByText(/processado|publicado|pronto para assistir/i)).toBeNull();

  releaseComplete(Response.json({ processing_status: "QUEUED" }, { status: 202 }));
  expect(await screen.findByText(/envio confirmado/i)).toBeDefined();
  expect(screen.queryByText(/publicado|pronto para assistir/i)).toBeNull();
  expect(localStorage.getItem("accessToken")).toBeNull();
  expect(localStorage.getItem("access_token")).toBeNull();
});

test("failed transfer does not claim the video was queued", async () => {
  const calls: string[] = [];
  stubSession((input) => {
    const url = String(input);
    calls.push(url);
    if (url.endsWith("/api/account/videos")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000aa",
        title: "Aula de Go",
        duration_seconds: 0,
        visibility: "PRIVATE",
        publication_status: "DRAFT",
        processing_status: "WAITING_UPLOAD",
      }, { status: 201 });
    }
    if (url.endsWith("/upload-sessions")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000bb",
        upload_url: "https://storage.example/videos/source",
        expires_at: "2026-10-14T20:00:00Z",
        max_bytes: 5,
      }, { status: 201 });
    }
    return Response.json({ error: { code: "NOT_FOUND", message: "rota não encontrada" } }, { status: 404 });
  });
  installXHR({ loaded: 1, total: 5, finish: false, fail: true });

  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );
  fireEvent.click(await screen.findByRole("button", { name: "Novo vídeo" }));
  fireEvent.change(await screen.findByRole("textbox", { name: "Título" }), {
    target: { value: "Aula de Go" },
  });
  fireEvent.change(screen.getByLabelText("Arquivo"), {
    target: { files: [new File(["hello"], "aula.mp4", { type: "video/mp4" })] },
  });
  fireEvent.click(screen.getByRole("button", { name: "Enviar" }));

  expect(await screen.findByText(/falha no envio/i)).toBeDefined();
  expect(screen.queryByText(/fila|enfileirado|processado|publicado/i)).toBeNull();
  expect(calls.some((url) => url.endsWith("/complete"))).toBe(false);
});

test("confirmed upload shows the queue and the running state", async () => {
  await renderConfirmed("QUEUED");
  expect(await screen.findByText("Na fila.")).toBeDefined();
  expect(screen.queryByText(/publicado|pronto para assistir/i)).toBeNull();
  cleanup();

  await renderConfirmed("PROCESSING");
  expect(await screen.findByText("Processando.")).toBeDefined();
  expect(screen.queryByText(/publicado|pronto para assistir/i)).toBeNull();
});

test("failed processing can be retried and ready does not offer playback", async () => {
  await renderConfirmed("FAILED", "o arquivo não tem vídeo");
  expect(await screen.findByText("o arquivo não tem vídeo")).toBeDefined();
  expect(screen.queryByText(/publicado|pronto para assistir/i)).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Tentar de novo" }));
  expect(await screen.findByText("Na fila.")).toBeDefined();
  cleanup();

  await renderConfirmed("READY");
  expect(await screen.findByText("Processamento concluído.")).toBeDefined();
  expect(screen.queryByText(/publicado|pronto para assistir/i)).toBeNull();
});

async function renderConfirmed(status: "QUEUED" | "PROCESSING" | "READY" | "FAILED", message?: string) {
  stubSession((input) => {
    const url = String(input);
    if (url.endsWith("/api/account/videos")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000aa",
        title: "Aula de Go",
        duration_seconds: 0,
        visibility: "PRIVATE",
        publication_status: "DRAFT",
        processing_status: "WAITING_UPLOAD",
      }, { status: 201 });
    }
    if (url.endsWith("/upload-sessions")) {
      return Response.json({
        id: "00000000-0000-4000-8000-0000000000bb",
        upload_url: "https://storage.example/videos/source",
        expires_at: "2026-10-14T20:00:00Z",
        max_bytes: 5,
      }, { status: 201 });
    }
    if (url.endsWith("/complete")) return Response.json({ processing_status: "QUEUED" }, { status: 202 });
    if (url.endsWith("/retry-processing")) return Response.json({ processing_status: "QUEUED" }, { status: 202 });
    if (url.endsWith("/processing")) {
      return Response.json({
        processing_status: status,
        error: message ? { code: "MEDIA_INVALID", message } : undefined,
      });
    }
    return Response.json({ error: { code: "NOT_FOUND", message: "rota não encontrada" } }, { status: 404 });
  });
  installXHR({ loaded: 5, total: 5, finish: true });
  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );
  fireEvent.click(await screen.findByRole("button", { name: "Novo vídeo" }));
  fireEvent.change(await screen.findByRole("textbox", { name: "Título" }), {
    target: { value: "Aula de Go" },
  });
  fireEvent.change(screen.getByLabelText("Arquivo"), {
    target: { files: [new File(["hello"], "aula.mp4", { type: "video/mp4" })] },
  });
  fireEvent.click(screen.getByRole("button", { name: "Enviar" }));
  expect(await screen.findByText(/envio confirmado/i)).toBeDefined();
}

function stubSession(next?: (input: RequestInfo | URL) => Promise<Response> | Response) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/api/auth/refresh") || url.endsWith("/api/auth/login")) {
        return Response.json({
          access_token: "access-token",
          token_type: "Bearer",
          expires_in: 900,
          user: { id: "user-1", email: "ada@example.com" },
        });
      }
      if (next) return next(input);
      return Response.json({ error: { code: "INVALID_REFRESH" } }, { status: 401 });
    }),
  );
}

let activeXHR: { finish: () => void } | null = null;

function installXHR(options: { loaded: number; total: number; finish: boolean; fail?: boolean }) {
  activeXHR = null;
  class FakeXHR {
    upload = new EventTarget();
    status = 0;
    method = "";
    url = "";
    listeners = new Map<string, Array<(event: Event) => void>>();

    open(method: string, url: string) {
      this.method = method;
      this.url = url;
    }

    setRequestHeader() {}

    addEventListener(type: string, listener: (event: Event) => void) {
      const list = this.listeners.get(type) ?? [];
      list.push(listener);
      this.listeners.set(type, list);
    }

    send() {
      activeXHR = {
        finish: () => {
          this.status = 200;
          this.emit("load");
        },
      };
      this.upload.dispatchEvent(new ProgressEvent("progress", {
        lengthComputable: true,
        loaded: options.loaded,
        total: options.total,
      }));
      if (options.fail) {
        this.emit("error");
        return;
      }
      if (options.finish) activeXHR?.finish();
    }

    emit(type: string) {
      for (const listener of this.listeners.get(type) ?? []) listener(new Event(type));
    }
  }

  vi.stubGlobal("XMLHttpRequest", FakeXHR);
}

function finishXHR() {
  activeXHR?.finish();
}
