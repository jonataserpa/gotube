import { z } from "zod";

export const publicVideoSchema = z.object({
  id: z.string().uuid(),
  title: z.string().min(1),
  duration_seconds: z.number().int().nonnegative(),
  thumbnail_url: z.string().url().nullable(),
  channel: z.object({
    name: z.string().min(1),
    handle: z.string().min(1),
  }),
});

export const catalogPageSchema = z.object({
  items: z.array(publicVideoSchema),
  next_cursor: z.string().min(1).nullable(),
});

export type PublicVideo = z.infer<typeof publicVideoSchema>;
