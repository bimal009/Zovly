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
export const PaginationQuerySchema = z
  .object({
    page: z.coerce.number().int().min(1).default(1),
    limit: z.coerce.number().int().min(1).max(100).default(10),
    search: z
      .string()
      .trim()
      .min(1)
      .max(255)
      .optional()
      .transform((value) => (value === "" ? undefined : value)),
  })
  .strict();

export type PaginationQuery = z.infer<typeof PaginationQuerySchema>;