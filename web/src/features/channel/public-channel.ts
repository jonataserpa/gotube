import { z } from "zod";

const publicChannelSchema = z.object({
  id: z.string().uuid(),
  handle: z.string().min(1),
  name: z.string().min(1),
  description: z.string(),
  categories: z.array(
    z.object({
      id: z.string().uuid(),
      slug: z.string().min(1),
      name: z.string().min(1),
    }),
  ),
});

export type PublicChannel = z.infer<typeof publicChannelSchema>;

export type PublicChannelState =
  | { kind: "ready"; channel: PublicChannel }
  | { kind: "missing" }
  | { kind: "error" };

export async function loadPublicChannel(handle: string): Promise<PublicChannelState> {
  const base = process.env.GO_API_INTERNAL_URL;
  if (!base) return { kind: "error" };
  let url: URL;
  try {
    url = new URL(`/v1/channels/${encodeURIComponent(handle)}`, base);
  } catch {
    return { kind: "error" };
  }
  try {
    const response = await fetch(url, { cache: "no-store" });
    if (response.status === 404) return { kind: "missing" };
    if (!response.ok) return { kind: "error" };
    const parsed = publicChannelSchema.safeParse(await response.json());
    if (!parsed.success) return { kind: "error" };
    return { kind: "ready", channel: parsed.data };
  } catch {
    return { kind: "error" };
  }
}
