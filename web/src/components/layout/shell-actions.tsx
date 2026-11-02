"use client";

import Link from "next/link";
import { useState } from "react";
import { AccountControl } from "@/components/layout/account-control";
import { NewVideoButton } from "@/features/upload/upload-dialog";
import { Icon, type IconName } from "@/components/layout/icon";
import { ThemeToggle } from "@/components/layout/theme-toggle";

const laterLabel = "Disponível numa etapa seguinte";

export function ShellActions() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <>
      <header className="fixed inset-x-0 top-0 z-50 border-b border-line/40 bg-background/80 backdrop-blur-xl">
        <div className="flex h-14 items-center justify-between gap-3 px-4 lg:px-6">
          <div className="flex min-w-0 items-center gap-3 lg:min-w-52">
            <button
              type="button"
              className="rounded-lg p-1 text-muted hover:bg-surface-high hover:text-foreground lg:hidden"
              aria-expanded={menuOpen}
              aria-controls="menu-lateral"
              onClick={() => setMenuOpen((open) => !open)}
            >
              <Icon name="menu" />
              <span className="sr-only">
                {menuOpen ? "Fechar menu" : "Abrir menu"}
              </span>
            </button>
            <Link href="/" className="text-base font-semibold tracking-tight">
              Go<span className="text-accent">Tube</span>
            </Link>
          </div>
          <form
            className="hidden min-w-0 flex-1 items-center justify-center gap-2 md:flex"
            role="search"
            onSubmit={(event) => event.preventDefault()}
          >
            <div className="relative w-full max-w-xl">
              <span className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted">
                <Icon name="search" />
              </span>
              <input
                className="h-9 w-full rounded-lg border border-line/50 bg-surface-lowest pr-14 pl-10 text-sm text-foreground placeholder:text-muted/70"
                placeholder="Pesquisar vídeos, canais e tópicos..."
                aria-label="Pesquisar vídeos, canais e tópicos"
                aria-describedby="busca-indisponivel"
                disabled
              />
              <kbd className="pointer-events-none absolute top-1/2 right-2 -translate-y-1/2 rounded border border-line bg-surface-high px-1.5 py-0.5 font-mono text-[11px] text-muted">
                ⌘K
              </kbd>
            </div>
            <p id="busca-indisponivel" className="sr-only">
              A pesquisa ainda não está disponível.
            </p>
            <button
              type="button"
              className="flex size-9 items-center justify-center rounded-lg border border-line/50 bg-surface-lowest text-muted"
              disabled
              aria-label={`Microfone. ${laterLabel}`}
            >
              <Icon name="mic" />
            </button>
          </form>
          <div className="flex items-center justify-end gap-2 lg:min-w-52">
            <NewVideoButton />
            <button
              type="button"
              className="rounded-lg p-1 text-muted disabled:cursor-not-allowed"
              disabled
              aria-label={`Notificações. ${laterLabel}`}
            >
              <Icon name="bell" />
            </button>
            <ThemeToggle />
            <AccountControl />
          </div>
        </div>
      </header>
      {menuOpen ? (
        <button
          type="button"
          className="fixed inset-0 top-14 z-30 bg-black/50 lg:hidden"
          aria-label="Fechar menu"
          onClick={() => setMenuOpen(false)}
        />
      ) : null}
      <aside
        id="menu-lateral"
        className={`fixed top-14 bottom-0 left-0 z-40 w-64 overflow-y-auto border-r border-line/40 bg-surface-lowest/95 backdrop-blur-xl ${menuOpen ? "translate-x-0" : "-translate-x-full"} lg:translate-x-0`}
      >
        <Sidebar />
      </aside>
    </>
  );
}

function Sidebar() {
  return (
    <div className="flex flex-col gap-5 p-2">
      <nav aria-label="Principal" className="flex flex-col gap-0.5">
        <Link
          href="/"
          aria-current="page"
          className="flex items-center gap-3 rounded-lg bg-surface-high px-3 py-2 text-sm font-semibold text-foreground"
        >
          <Icon name="home" />
          Início
        </Link>
        <UnavailableItem icon="bolt" label="Shorts" />
        <UnavailableItem icon="subscriptions" label="Inscrições" />
      </nav>
      <div className="h-px bg-line/40" />
      <nav aria-label="Biblioteca pessoal" className="flex flex-col gap-0.5">
        <p className="px-3 py-1 font-mono text-xs tracking-wider text-muted uppercase">
          Biblioteca pessoal
        </p>
        <UnavailableItem icon="library" label="Biblioteca" />
        <UnavailableItem icon="history" label="Histórico" />
        <UnavailableItem icon="videos" label="Seus vídeos" />
        <UnavailableItem icon="clock" label="Assistir mais tarde" />
        <UnavailableItem icon="playlist" label="Playlists" />
      </nav>
      <div className="h-px bg-line/40" />
      <section aria-labelledby="inscricoes-titulo" className="px-3 py-1">
        <h2
          id="inscricoes-titulo"
          className="font-mono text-xs tracking-wider text-muted uppercase"
        >
          Inscrições
        </h2>
        <p className="mt-2 text-sm text-muted">Nenhuma inscrição ainda.</p>
      </section>
    </div>
  );
}

function UnavailableItem({
  icon,
  label,
}: {
  icon: IconName;
  label: string;
}) {
  return (
    <button
      type="button"
      className="flex items-center gap-3 rounded-lg px-3 py-2 text-left text-sm text-muted disabled:cursor-not-allowed"
      disabled
      aria-label={`${label}. Disponível numa etapa seguinte`}
    >
      <Icon name={icon} />
      {label}
    </button>
  );
}
