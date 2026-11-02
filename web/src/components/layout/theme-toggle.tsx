"use client";

import { toggleDocumentTheme } from "@/components/layout/theme";

export function ThemeToggle() {
  return (
    <button
      type="button"
      className="rounded-lg px-2 py-1 text-sm text-muted hover:bg-surface-high hover:text-foreground"
      aria-label="Alternar tema claro e escuro"
      onClick={() => {
        toggleDocumentTheme(document.documentElement, localStorage);
      }}
    >
      Tema
    </button>
  );
}
