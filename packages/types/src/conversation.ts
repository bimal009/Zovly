
import { z } from "zod";

export const createConversationSchema = z.object({

  platform: z.enum([
    "instagram",
    "facebook",
    "whatsapp",
    "tiktok",
    "web"
  ]),

  threadId: z.string().min(1),
  contactId: z.string().min(1),

  contactName: z.string().optional(),
  contactUsername: z.string().optional(),
  contactAvatarUrl: z.string().url().optional(),
});

export type CreateConversationInput = z.infer<typeof createConversationSchema>;