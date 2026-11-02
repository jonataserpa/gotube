import type { ReactNode } from "react";
import { ShellActions } from "@/components/layout/shell-actions";

export function SiteShell({ children }: { children: ReactNode }) {
  return (
    <>
      <a href="#conteudo" className="skip-link">
        Pular para o conteúdo
      </a>
      <ShellActions />
      <div className="lg:pl-64">
        <main id="conteudo" className="min-h-screen px-4 pt-14 pb-10 lg:px-6">
          {children}
        </main>
      </div>
    </>
  );
}
