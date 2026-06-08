import {
  pgTable,
  pgEnum,
  uuid,
  text,
  integer,
  timestamp,
  index,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";
import { plans, billingCycleEnum } from "./plans";
import { businessSubscriptions } from "./subscriptions";

// ── Enums ─────────────────────────────────────────────────────────────────────

export const paymentStatusEnum = pgEnum("payment_status", [
  "paid",
  "refunded",
  "partially_refunded",
  "failed",
]);

// ── Payment Records ───────────────────────────────────────────────────────────
// One row per Paddle transaction. Populated via Paddle webhooks.

export const paymentRecords = pgTable(
  "payment_records",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull(),

    subscriptionId: uuid("subscription_id").references(
      () => businessSubscriptions.id,
      { onDelete: "set null" },
    ),

    planId: uuid("plan_id").references(() => plans.id),
    billingCycle: billingCycleEnum("billing_cycle").notNull(),

    // ── Paddle transaction data ───────────────────────────────────────────────
    paddleTransactionId: text("paddle_transaction_id").unique().notNull(), // txn_xxxx
    paddleSubscriptionId: text("paddle_subscription_id"),                  // sub_xxxx
    paddleCustomerId: text("paddle_customer_id"),                          // ctm_xxxx

    // ── Amount (minor units, e.g. cents for USD) ──────────────────────────────
    amount: integer("amount").notNull(),
    currency: text("currency").notNull().default("USD"),

    periodStart: timestamp("period_start").notNull(),
    periodEnd: timestamp("period_end").notNull(),

    status: paymentStatusEnum("status").notNull().default("paid"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("payment_business_idx").on(table.businessId),
    index("payment_status_idx").on(table.status),
    index("payment_paddle_txn_idx").on(table.paddleTransactionId),
    index("payment_paddle_sub_idx").on(table.paddleSubscriptionId),
    index("payment_sub_idx").on(table.subscriptionId),
  ],
);

// ── Relations ─────────────────────────────────────────────────────────────────

export const paymentRecordRelations = relations(paymentRecords, ({ one }) => ({
  business: one(business, {
    fields: [paymentRecords.businessId],
    references: [business.id],
  }),
  subscription: one(businessSubscriptions, {
    fields: [paymentRecords.subscriptionId],
    references: [businessSubscriptions.id],
  }),
  plan: one(plans, {
    fields: [paymentRecords.planId],
    references: [plans.id],
  }),
}));

// ── Type helpers ──────────────────────────────────────────────────────────────

export type PaymentRecord = typeof paymentRecords.$inferSelect;
export type NewPaymentRecord = typeof paymentRecords.$inferInsert;
export type PaymentStatus = (typeof paymentStatusEnum.enumValues)[number];
