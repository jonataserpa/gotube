import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { HomePage } from "@/components/catalog/home-page";
import { SiteShell } from "@/components/layout/site-shell";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

test("signed out shell offers Entrar and keeps later actions disabled", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => Response.json({ error: { code: "INVALID_REFRESH" } }, { status: 401 })),
  );

  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );

  expect(await screen.findByRole("link", { name: "Entrar" })).toHaveProperty(
    "pathname",
    "/login",
  );
  expect(screen.getByRole("button", { name: /Novo vídeo/ })).toHaveProperty(
    "disabled",
    true,
  );
  expect(
    screen.getByRole("textbox", { name: "Pesquisar vídeos, canais e tópicos" }),
  ).toHaveProperty("disabled", true);
  expect(screen.queryByRole("button", { name: /Conta/ })).toBeNull();
});

test("restored session shows the account email and logout", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      Response.json({
        access_token: "access-token",
        token_type: "Bearer",
        expires_in: 900,
        user: { id: "user-1", email: "ada@example.com" },
      }),
    ),
  );

  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );

  expect(await screen.findByText("ada@example.com")).toBeDefined();
  expect(screen.getByRole("button", { name: "Sair" })).toBeDefined();
  expect(screen.getByRole("link", { name: "Perfil" }).getAttribute("href")).toBe("/perfil");
  expect(screen.getByRole("button", { name: "Novo vídeo" })).toHaveProperty("disabled", false);
  expect(
    screen.getByRole("textbox", { name: "Pesquisar vídeos, canais e tópicos" }),
  ).toHaveProperty("disabled", true);
  expect(screen.queryByRole("link", { name: "Entrar" })).toBeNull();
});

test("catalog card does not show the account email", () => {
  render(
    <HomePage
      state={{
        kind: "ready",
        items: [
          {
            id: "00000000-0000-4000-8000-000000000010",
            title: "Catálogo público de exemplo",
            duration_seconds: 125,
            thumbnail_url: null,
            channel: { name: "Laboratório GoTube", handle: "laboratorio" },
          },
        ],
        nextCursor: null,
      }}
    />,
  );

  expect(screen.queryByText("ada@example.com")).toBeNull();
  expect(screen.getByText("Laboratório GoTube")).toBeDefined();
});
