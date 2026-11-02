"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { AuthPanel, authFieldClass } from "@/features/identity/auth-panel";
import { requestPasswordReset } from "@/features/identity/session";

export function ForgotForm() {
  const [message, setMessage] = useState<string | null>(null);
  const [resetToken, setResetToken] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPending(true);
    const result = await requestPasswordReset(String(data.get("email") ?? ""));
    setPending(false);
    if (!result) {
      setMessage("Não foi possível registrar a recuperação.");
      setResetToken(null);
      return;
    }
    setMessage(result.message);
    setResetToken(result.resetToken);
  }

  return (
    <AuthPanel title="Recuperar senha">
      <form onSubmit={onSubmit} className="flex flex-col gap-3">
        <label className="flex flex-col gap-1 text-sm">
          E-mail
          <input className={authFieldClass} name="email" type="email" autoComplete="email" required />
        </label>
        {message ? <p className="text-sm leading-6">{message}</p> : null}
        {resetToken ? (
          <Link href={`/reset-password?token=${encodeURIComponent(resetToken)}`} className="text-sm text-accent">
            Redefinir senha neste laboratório
          </Link>
        ) : null}
        <button
          type="submit"
          className="h-10 rounded-lg bg-accent text-sm font-medium text-white disabled:opacity-70"
          disabled={pending}
        >
          {pending ? "Enviando..." : "Pedir recuperação"}
        </button>
      </form>
    </AuthPanel>
  );
}
