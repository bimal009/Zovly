import {
  pgTable,
  uuid,
  text,
  integer,
  timestamp,
  index,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";
import { plans } from "./plans";
import { businessSubscriptions } from "./subscriptions";
import { billingCycleEnum, paymentStatusEnum } from "./enums"; // ← from enums

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

    paddleTransactionId: text("paddle_transaction_id").unique().notNull(),
    paddleSubscriptionId: text("paddle_subscription_id"),
    paddleCustomerId: text("paddle_customer_id"),

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

export type PaymentRecord = typeof paymentRecords.$inferSelect;
export type NewPaymentRecord = typeof paymentRecords.$inferInsert;
export type PaymentStatus = (typeof paymentStatusEnum.enumValues)[number];
