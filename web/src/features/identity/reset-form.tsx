"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { AuthPanel, authFieldClass } from "@/features/identity/auth-panel";
import { resetPassword } from "@/features/identity/session";

export function ResetForm({ token }: { token: string }) {
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPending(true);
    const message = await resetPassword(token, String(data.get("password") ?? ""));
    setPending(false);
    if (message) {
      setError(message);
      return;
    }
    setDone(true);
  }

  return (
    <AuthPanel title="Redefinir senha">
      {token === "" ? (
        <p className="text-sm leading-6">O token de recuperação não veio no endereço.</p>
      ) : done ? (
        <p className="text-sm leading-6">
          Senha atualizada.{" "}
          <Link href="/login" className="text-accent">
            Entrar
          </Link>
        </p>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <label className="flex flex-col gap-1 text-sm">
            Nova senha
            <input
              className={authFieldClass}
              name="password"
              type="password"
              autoComplete="new-password"
              minLength={8}
              required
            />
          </label>
          {error ? <p className="text-sm text-accent">{error}</p> : null}
          <button
            type="submit"
            className="h-10 rounded-lg bg-accent text-sm font-medium text-white disabled:opacity-70"
            disabled={pending}
          >
            {pending ? "Salvando..." : "Salvar senha"}
          </button>
        </form>
      )}
    </AuthPanel>
  );
}
