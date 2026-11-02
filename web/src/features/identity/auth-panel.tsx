import type { ReactNode } from "react";

export function AuthPanel({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="mx-auto flex w-full max-w-md flex-col gap-4 py-10">
      <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
      {children}
    </section>
  );
}

export const authFieldClass =
  "h-10 rounded-lg border border-line/70 bg-card px-3 text-sm text-foreground";
