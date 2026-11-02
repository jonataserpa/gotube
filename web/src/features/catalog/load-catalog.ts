import { catalogPageSchema } from "@/features/catalog/catalog-schema";
import type { PublicVideo } from "@/features/catalog/catalog-schema";

export type CatalogState =
  | { kind: "ready"; items: PublicVideo[]; nextCursor: string | null }
  | { kind: "empty" }
  | { kind: "error"; retryHref: string };

export function catalogRetryHref(cursor: string | undefined): string {
  if (!cursor) {
    return "/";
  }
  return `/?cursor=${encodeURIComponent(cursor)}`;
}

export async function loadCatalog(cursor: string | undefined): Promise<CatalogState> {
  const retryHref = catalogRetryHref(cursor);
  const base = process.env.GO_API_INTERNAL_URL;
  if (!base) {
    return { kind: "error", retryHref };
  }
  let url: URL;
  try {
    url = new URL("/v1/videos", base);
  } catch {
    return { kind: "error", retryHref };
  }
  url.searchParams.set("limit", "12");
  if (cursor) {
    url.searchParams.set("cursor", cursor);
  }
  try {
    const response = await fetch(url, { cache: "no-store" });
    if (!response.ok) {
      return { kind: "error", retryHref };
    }
    const parsed = catalogPageSchema.safeParse(await response.json());
    if (!parsed.success) {
      return { kind: "error", retryHref };
    }
    if (parsed.data.items.length === 0) {
      return { kind: "empty" };
    }
    return {
      kind: "ready",
      items: parsed.data.items,
      nextCursor: parsed.data.next_cursor,
    };
  } catch {
    return { kind: "error", retryHref };
  }
}
