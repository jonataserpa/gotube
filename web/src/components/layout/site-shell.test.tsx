import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { SiteShell } from "./site-shell";

test("shell exposes skip link, navigation and main landmark", () => {
  render(
    <SiteShell>
      <p>conteúdo</p>
    </SiteShell>,
  );

  expect(
    screen.getByRole("link", { name: "Pular para o conteúdo" }),
  ).toHaveProperty("hash", "#conteudo");
  expect(screen.getByRole("navigation", { name: "Principal" })).toBeDefined();
  expect(screen.getByRole("link", { name: "Início" })).toHaveProperty(
    "ariaCurrent",
    "page",
  );
  expect(screen.getByRole("main")).toBeDefined();
  expect(
    screen.getByRole("button", { name: "Alternar tema claro e escuro" }),
  ).toBeDefined();
  expect(
    screen.getByRole("button", { name: /Novo vídeo/ }),
  ).toHaveProperty("disabled", true);
  expect(screen.queryByText("NextTube")).toBeNull();
  expect(screen.getByText("Nenhuma inscrição ainda.")).toBeDefined();
});
