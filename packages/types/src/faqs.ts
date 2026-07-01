import { z } from "zod";

export const createFaqSchema = z
  .object({
    question: z
      .string()
      .trim()
      .min(1, "Question is required")
      .max(150, "Question is too long"),

    answer: z
      .string()
      .trim()
      .min(1, "Answer is required")
      .max(300, "Answer must be at most 300 characters"),

    isActive: z.boolean(),
  })
  .strict();

export const updateFaqSchema = createFaqSchema.partial().strict();

export type CreateFaqInput = z.infer<typeof createFaqSchema>;
export type UpdateFaqInput = z.infer<typeof updateFaqSchema>;

export type Faq = {
  id: string;
  businessId: string;
  question: string;
  answer: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: Date;
  updatedAt: Date;
};