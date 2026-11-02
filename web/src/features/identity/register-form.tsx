"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { AuthPanel, authFieldClass } from "@/features/identity/auth-panel";
import { registerAccount } from "@/features/identity/session";

export function RegisterForm() {
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState(false);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPending(true);
    const message = await registerAccount(
      String(data.get("email") ?? ""),
      String(data.get("password") ?? ""),
    );
    setPending(false);
    if (message) {
      setError(message);
      return;
    }
    setCreated(true);
    setError(null);
  }

  return (
    <AuthPanel title="Criar conta">
      {created ? (
        <p className="text-sm leading-6">
          Conta criada.{" "}
          <Link href="/login" className="text-accent">
            Entre com seu e-mail.
          </Link>
        </p>
      ) : (
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <label className="flex flex-col gap-1 text-sm">
            E-mail
            <input className={authFieldClass} name="email" type="email" autoComplete="email" required />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            Senha
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
            {pending ? "Criando..." : "Cadastrar"}
          </button>
        </form>
      )}
    </AuthPanel>
  );
}
