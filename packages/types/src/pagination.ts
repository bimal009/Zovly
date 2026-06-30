import { z } from "zod";

export const MetaSchema = z
  .object({
    page: z.number().int().min(1).optional(),
    limit: z.number().int().min(1).max(100).optional(),
    total: z.number().int().min(0).optional(),
    totalPages: z.number().int().min(0).optional(),
  })
  .strict();

export type Meta = z.infer<typeof MetaSchema>;
