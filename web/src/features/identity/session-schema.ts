import { z } from "zod";

export const sessionSchema = z.object({
  access_token: z.string().min(1),
  token_type: z.literal("Bearer"),
  expires_in: z.number(),
  user: z.object({
    id: z.string().min(1),
    email: z.string().min(1),
  }),
});

export const forgotResultSchema = z.object({
  message: z.string().min(1),
  reset_token: z.string().min(1).optional(),
});

export const apiErrorSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
  }),
});

export type AccountSession = {
  accessToken: string;
  email: string;
  userId: string;
};
