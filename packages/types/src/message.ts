import { z } from "zod";

export const createMessageSchema = z.object({
  direction: z.enum(["in", "out"]),
  content: z.string().min(1).max(10000).optional(),
  mediaUrl: z.url().optional(),
  mediaType: z.enum(["image", "video", "audio", "document", "link"]).optional(),
  platformMessageId: z.string().optional(),
  replyToMessageId: z.string().uuid().optional(),
  platformSenderId: z.string().optional(),
}).refine((data) => data.content || data.mediaUrl, {
  message: "Either content or mediaUrl must be provided",
});

export type CreateMessageInput = z.infer<typeof createMessageSchema>;
