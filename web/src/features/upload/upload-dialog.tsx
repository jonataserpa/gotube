"use client";

import { useEffect, useState, useSyncExternalStore, type FormEvent } from "react";
import {
  getSession,
  restoreSession,
  subscribe,
} from "@/features/identity/session";
import { Icon } from "@/components/layout/icon";
import {
  completeUpload,
  createDraft,
  createUploadSession,
  readProcessing,
  retryProcessing,
  sendBytes,
  type ProcessingView,
} from "@/features/upload/direct-upload";

const laterLabel = "Disponível numa etapa seguinte";

export function NewVideoButton() {
  const session = useSyncExternalStore(subscribe, getSession, () => null);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    void restoreSession();
  }, []);

  if (!session) {
    return (
      <button
        type="button"
        className="flex items-center gap-1 rounded-lg bg-accent px-3 py-1.5 text-[13px] font-medium text-white disabled:cursor-not-allowed disabled:opacity-70"
        disabled
        aria-label={`Novo vídeo. ${laterLabel}`}
      >
        <Icon name="add" />
        <span className="hidden sm:inline">Novo vídeo</span>
      </button>
    );
  }

  return (
    <>
      <button
        type="button"
        className="flex items-center gap-1 rounded-lg bg-accent px-3 py-1.5 text-[13px] font-medium text-white"
        aria-label="Novo vídeo"
        onClick={() => setOpen(true)}
      >
        <Icon name="add" />
        <span className="hidden sm:inline">Novo vídeo</span>
      </button>
      {open ? (
        <UploadDialog token={session.accessToken} onClose={() => setOpen(false)} />
      ) : null}
    </>
  );
}

function ProcessingLine({ view, onRetry }: { view: ProcessingView; onRetry: () => void }) {
  const label = processingLabel(view);
  return (
    <div className="mt-3 text-sm" aria-live="polite">
      <p>{label}</p>
      {view.processing_status === "FAILED" ? (
        <button type="button" className="mt-3 rounded-lg border border-line px-3 py-1.5 text-sm" onClick={onRetry}>
          Tentar de novo
        </button>
      ) : null}
    </div>
  );
}

function processingLabel(view: ProcessingView): string {
  if (view.processing_status === "QUEUED") return "Na fila.";
  if (view.processing_status === "PROCESSING") return "Processando.";
  if (view.processing_status === "READY") return "Processamento concluído.";
  if (view.processing_status === "FAILED") {
    return view.error?.message || "Não foi possível processar o vídeo.";
  }
  return "Aguardando processamento.";
}

function UploadDialog({ token, onClose }: { token: string; onClose: () => void }) {
  const [title, setTitle] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [phase, setPhase] = useState<"form" | "sending" | "received" | "confirmed" | "failed">("form");
  const [percent, setPercent] = useState(0);
  const [notice, setNotice] = useState("");
  const [videoId, setVideoId] = useState("");
  const [processing, setProcessing] = useState<ProcessingView | null>(null);

  async function submit(event: FormEvent) {
    event.preventDefault();
    const trimmed = title.trim();
    if (!trimmed || !file) {
      setNotice("Informe o título e escolha um arquivo.");
      return;
    }
    setNotice("");
    setPhase("sending");
    setPercent(0);
    try {
      const createdId = await createDraft(token, trimmed);
      setVideoId(createdId);
      const session = await createUploadSession(token, createdId, file.size);
      await sendBytes(session.uploadUrl, file, (next) => {
        setPercent(next);
        if (next === 100) setPhase("received");
      });
      setPhase("received");
      await completeUpload(token, session.id);
      setPhase("confirmed");
    } catch {
      setPhase("failed");
    }
  }

  useEffect(() => {
    if (phase !== "confirmed" || !videoId) return;
    let stop = false;
    async function poll() {
      try {
        const view = await readProcessing(token, videoId);
        if (!stop && view) setProcessing(view);
      } catch {
        // A resposta atrasada não substitui o estado já mostrado.
      }
    }
    void poll();
    const timer = window.setInterval(() => void poll(), 2000);
    return () => {
      stop = true;
      window.clearInterval(timer);
    };
  }, [phase, token, videoId]);

  async function retry() {
    if (!videoId) return;
    await retryProcessing(token, videoId);
    setProcessing({ processing_status: "QUEUED" });
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-end justify-center bg-black/60 p-4 sm:items-center" role="presentation">
      <form
        role="dialog"
        aria-labelledby="novo-video-titulo"
        className="w-full max-w-md rounded-xl border border-line bg-surface-lowest p-4"
        onSubmit={(event) => void submit(event)}
      >
        <h2 id="novo-video-titulo" className="text-base font-semibold">
          Novo vídeo
        </h2>
        <label className="mt-4 block text-sm" htmlFor="titulo-video">
          Título
          <input
            id="titulo-video"
            className="mt-1 h-9 w-full rounded-lg border border-line/50 bg-background px-3 text-sm"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />
        </label>
        <label className="mt-3 block text-sm" htmlFor="arquivo-video">
          Arquivo
          <input
            id="arquivo-video"
            className="mt-1 block w-full text-sm"
            type="file"
            accept="video/*"
            onChange={(event) => setFile(event.target.files?.[0] ?? null)}
          />
        </label>
        {phase === "sending" || phase === "received" ? (
          <div
            className="mt-4"
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={percent}
            aria-label="Progresso do envio"
          >
            {phase === "received" ? "Arquivo recebido." : `Enviando ${percent}%`}
          </div>
        ) : null}
        {phase === "confirmed" ? <p className="mt-4 text-sm">Envio confirmado.</p> : null}
        {phase === "confirmed" && processing ? (
          <ProcessingLine view={processing} onRetry={() => void retry()} />
        ) : null}
        {phase === "failed" ? <p className="mt-4 text-sm">Falha no envio.</p> : null}
        {notice ? <p className="mt-4 text-sm">{notice}</p> : null}
        <div className="mt-4 flex justify-end gap-2">
          <button type="button" className="rounded-lg border border-line px-3 py-1.5 text-sm" onClick={onClose}>
            Fechar
          </button>
          <button
            type="submit"
            className="rounded-lg bg-accent px-3 py-1.5 text-sm text-white disabled:opacity-70"
            disabled={phase === "sending" || phase === "received" || phase === "confirmed"}
          >
            Enviar
          </button>
        </div>
      </form>
    </div>
  );
}
