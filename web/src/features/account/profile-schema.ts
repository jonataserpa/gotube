import { z } from "zod";

export const profileSchema = z.object({
  id: z.string().min(1),
  email: z.string().min(1),
  display_name: z.string().nullable(),
});

export const categorySchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(1),
  name: z.string().min(1),
});

export const channelDocumentSchema = z.object({
  id: z.string().uuid(),
  handle: z.string().min(1),
  name: z.string().min(1),
  description: z.string(),
  categories: z.array(categorySchema),
});

export const ownChannelSchema = z.object({
  channel: channelDocumentSchema.nullable(),
});

export const categoryPageSchema = z.object({
  items: z.array(categorySchema),
});

export type Profile = z.infer<typeof profileSchema>;
export type OwnChannel = z.infer<typeof ownChannelSchema>;
export type Category = z.infer<typeof categorySchema>;
