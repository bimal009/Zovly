import { pgEnum } from "drizzle-orm/pg-core";

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
