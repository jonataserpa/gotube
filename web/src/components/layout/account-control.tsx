"use client";

import Link from "next/link";
import { useEffect, useSyncExternalStore } from "react";
import {
  getSession,
  logout,
  restoreSession,
  subscribe,
} from "@/features/identity/session";

export function AccountControl() {
  const session = useSyncExternalStore(subscribe, getSession, () => null);

  useEffect(() => {
    void restoreSession();
  }, []);

  if (!session) {
    return (
      <Link
        href="/login"
        className="rounded-lg border border-line px-3 py-1.5 text-[13px] font-medium text-foreground"
      >
        Entrar
      </Link>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <span className="max-w-36 truncate text-xs text-muted">{session.email}</span>
      <Link
        href="/perfil"
        className="rounded-lg border border-line px-3 py-1.5 text-[13px] font-medium text-foreground"
      >
        Perfil
      </Link>
      <button
        type="button"
        className="rounded-lg border border-line px-3 py-1.5 text-[13px] font-medium text-foreground"
        onClick={() => void logout()}
      >
        Sair
      </button>
    </div>
  );
}
