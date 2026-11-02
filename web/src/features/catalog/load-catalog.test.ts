import { expect, test, vi } from "vitest";
import { loadCatalog } from "./load-catalog";

const item = {
  id: "00000000-0000-4000-8000-000000000010",
  title: "Catálogo público de exemplo",
  duration_seconds: 125,
  thumbnail_url: null,
  channel: { name: "Laboratório GoTube", handle: "laboratorio" },
};

test("missing API URL is an error, not an empty catalog", async () => {
  vi.stubEnv("GO_API_INTERNAL_URL", "");
  await expect(loadCatalog(undefined)).resolves.toEqual({
    kind: "error",
    retryHref: "/",
  });
  vi.unstubAllEnvs();
});

test("invalid payload is an error", async () => {
  vi.stubEnv("GO_API_INTERNAL_URL", "http://127.0.0.1:8080");
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response(JSON.stringify({ items: "nope" }), { status: 200 })),
  );
  await expect(loadCatalog("abc")).resolves.toEqual({
    kind: "error",
    retryHref: "/?cursor=abc",
  });
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

test("an empty page stays empty", async () => {
  vi.stubEnv("GO_API_INTERNAL_URL", "http://127.0.0.1:8080");
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      new Response(JSON.stringify({ items: [], next_cursor: null }), { status: 200 }),
    ),
  );
  await expect(loadCatalog(undefined)).resolves.toEqual({ kind: "empty" });
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

test("a public page keeps the next cursor", async () => {
  vi.stubEnv("GO_API_INTERNAL_URL", "http://127.0.0.1:8080");
  const fetchMock = vi.fn(
    async () =>
      new Response(JSON.stringify({ items: [item], next_cursor: "next" }), { status: 200 }),
  );
  vi.stubGlobal("fetch", fetchMock);
  await expect(loadCatalog("prev")).resolves.toEqual({
    kind: "ready",
    items: [item],
    nextCursor: "next",
  });
  const calls = fetchMock.mock.calls as unknown[][];
  const called = String(calls[0]?.[0]);
  expect(called).toContain("/v1/videos");
  expect(called).toContain("cursor=prev");
  expect(called).not.toContain("NEXT_PUBLIC_");
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});
