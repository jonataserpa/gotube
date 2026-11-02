"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { AuthPanel, authFieldClass } from "@/features/identity/auth-panel";
import { login } from "@/features/identity/session";

export function LoginForm() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setPending(true);
    const message = await login(
      String(data.get("email") ?? ""),
      String(data.get("password") ?? ""),
    );
    setPending(false);
    if (message) {
      setError(message);
      return;
    }
    router.push("/");
  }

  return (
    <AuthPanel title="Entrar">
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
            autoComplete="current-password"
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
          {pending ? "Entrando..." : "Entrar"}
        </button>
      </form>
      <p className="text-sm text-muted">
        <Link href="/register" className="text-foreground">
          Criar conta
        </Link>
        {" · "}
        <Link href="/forgot-password" className="text-foreground">
          Esqueci minha senha
        </Link>
      </p>
    </AuthPanel>
  );
}
