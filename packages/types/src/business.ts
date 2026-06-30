import z from "zod";

export const createBusinessSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Business name is required")
      .max(100, "Business name must be at most 100 characters"),

    description: z
      .string()
      .trim()
      .min(1, "Description is required")
      .min(10, "Description must be at least 10 characters")
      .max(2000, "Description must be at most 2000 characters"),

    logo: z.url().optional(),
    website: z.url().optional(),

    phone: z
      .string()
      .trim()
      .min(3, "Phone number is too short")
      .max(30, "Phone number is too long"),

    address: z
      .string()
      .trim()
      .min(1, "Address is required")
      .min(5, "Address must be at least 5 characters")
      .max(255, "Address must be at most 255 characters"),

    city: z
      .string()
      .trim()
      .min(1, "City is required")
      .max(100, "City must be at most 100 characters"),

    country: z
      .string()
      .trim()
      .length(3, "Country must be a 3-letter ISO code")
      .default("NPL"),

    type: z.enum(["product", "service", "both"]).default("service"),
  })
  .strict();

export const updateBusinessSchema = createBusinessSchema.partial().strict();

export type CreateBusinessInput = z.infer<typeof createBusinessSchema>;
export type UpdateBusinessInput = z.infer<typeof updateBusinessSchema>;
