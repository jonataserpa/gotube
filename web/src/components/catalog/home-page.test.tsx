import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";

afterEach(() => {
  cleanup();
});
import { HomePage } from "./home-page";

const video = {
  id: "00000000-0000-4000-8000-000000000010",
  title: "Catálogo público de exemplo",
  duration_seconds: 125,
  thumbnail_url: "https://example.com/gotube/dev-thumb.jpg" as string | null,
  channel: { name: "Laboratório GoTube", handle: "laboratorio" },
};

test("grid shows title, duration, channel and thumbnail", () => {
  render(
    <HomePage
      state={{
        kind: "ready",
        items: [video],
        nextCursor: "cursor-1",
      }}
    />,
  );

  expect(
    screen.getByRole("heading", { level: 2, name: video.title }),
  ).toBeDefined();
  expect(screen.getByText("Laboratório GoTube")).toBeDefined();
  expect(screen.getByText("2:05")).toBeDefined();
  expect(
    screen.getByRole("img", { name: "Miniatura de Catálogo público de exemplo" }),
  ).toBeDefined();
  expect(screen.getByRole("link", { name: "Próxima página" }).getAttribute("href")).toBe(
    "/?cursor=cursor-1",
  );
  expect(screen.queryByRole("link", { name: video.title })).toBeNull();
  expect(screen.queryByText(/ainda não está implementado/i)).toBeNull();
});

test("channel name links to the public page and the card does not play", () => {
  render(
    <HomePage
      state={{
        kind: "ready",
        items: [video],
        nextCursor: null,
      }}
    />,
  );

  expect(screen.getByRole("link", { name: "Laboratório GoTube" }).getAttribute("href")).toBe(
    "/channels/laboratorio",
  );
  for (const link of screen.getAllByRole("link")) {
    const href = link.getAttribute("href") ?? "";
    expect(href.includes("/watch") || href.includes("/playback")).toBe(false);
  }
});

test("missing thumbnail reserves an accessible area", () => {
  render(
    <HomePage
      state={{
        kind: "ready",
        items: [{ ...video, thumbnail_url: null }],
        nextCursor: null,
      }}
    />,
  );

  expect(screen.getByRole("img", { name: "Sem miniatura" })).toBeDefined();
  expect(screen.queryByRole("link", { name: "Próxima página" })).toBeNull();
});

test("empty catalog is distinct from the old unimplemented copy", () => {
  render(<HomePage state={{ kind: "empty" }} />);

  expect(screen.getByText(/ainda não há vídeos públicos/i)).toBeDefined();
  expect(screen.queryByText(/ainda não está implementado/i)).toBeNull();
});

test("api error offers retry and does not look like an empty list", () => {
  render(<HomePage state={{ kind: "error", retryHref: "/?cursor=abc" }} />);

  expect(screen.getByText(/não foi possível carregar o catálogo/i)).toBeDefined();
  expect(screen.getByRole("link", { name: "Tentar de novo" }).getAttribute("href")).toBe(
    "/?cursor=abc",
  );
  expect(screen.queryByText(/ainda não há vídeos públicos/i)).toBeNull();
  expect(screen.queryByText(/ainda não está implementado/i)).toBeNull();
});
