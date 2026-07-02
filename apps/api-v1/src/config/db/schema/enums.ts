import { pgEnum } from "drizzle-orm/pg-core";

export const platformEnum = pgEnum("platform", [
  "instagram",
  "facebook",
  "whatsapp",
  "tiktok",
  "web"
]);

export const messageDirectionEnum = pgEnum("message_direction", ["in", "out"]);

export const messageSenderEnum = pgEnum("message_sender", ["ai", "human"]);

export const messageMediaTypeEnum = pgEnum("message_media_type", [
  "image",
  "video",
  "audio",
  "document",
  "link",
]);

export const billingCycleEnum = pgEnum("billing_cycle", ["monthly", "yearly"]);

export const planStatusEnum = pgEnum("plan_status", [
  "active",
  "trialing",
  "past_due",
  "paused",
  "cancelled",
  "expired",
]);

export const paymentStatusEnum = pgEnum("payment_status", [
  "paid",
  "refunded",
  "partially_refunded",
  "failed",
]);

export const paymentMethodEnum = pgEnum("payment_method", [
  "esewa",
  "khalti",
  "fonepay",
  "bank_transfer",
  "cash",
  "other",
]);
// enums.ts
export const messageStatusEnum = pgEnum("message_status", [
  "pending", // outbound, generated, not yet sent
  "sent", // Graph API accepted (200)
  "failed", // Graph API rejected — see error
  "skipped", // not sendable (e.g. 24h window closed)
]);
