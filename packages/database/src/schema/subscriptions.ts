import {
  pgTable,
  uuid,
  text,
  integer,
  boolean,
  timestamp,
  index,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";
import { plans } from "./plans";
import { paymentRecords } from "./payments";
import { billingCycleEnum, planStatusEnum } from "./enums"; // ← from enums

export const businessSubscriptions = pgTable(
  "business_subscriptions",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull()
      .unique(),

    planId: uuid("plan_id")
      .references(() => plans.id)
      .notNull(),

    paddleSubscriptionId: text("paddle_subscription_id").unique(),
    paddleCustomerId: text("paddle_customer_id"),
    paddlePriceId: text("paddle_price_id"),

    billingCycle: billingCycleEnum("billing_cycle")
      .notNull()
      .default("monthly"),
    status: planStatusEnum("status").notNull().default("trialing"),

    aiRepliesUsed: integer("ai_replies_used").default(0).notNull(),
    postsUsed: integer("posts_used").default(0).notNull(),
    usageResetAt: timestamp("usage_reset_at"),

    trialStartedAt: timestamp("trial_started_at"),
    trialEndsAt: timestamp("trial_ends_at"),

    currentPeriodStart: timestamp("current_period_start"),
    currentPeriodEnd: timestamp("current_period_end"),
    cancelAtPeriodEnd: boolean("cancel_at_period_end").default(false).notNull(),
    cancelledAt: timestamp("cancelled_at"),
    pausedAt: timestamp("paused_at"),

    notes: text("notes"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("sub_business_idx").on(table.businessId),
    index("sub_plan_idx").on(table.planId),
    index("sub_status_idx").on(table.status),
    index("sub_paddle_sub_idx").on(table.paddleSubscriptionId),
    index("sub_paddle_customer_idx").on(table.paddleCustomerId),
  ],
);

export const businessSubscriptionRelations = relations(
  businessSubscriptions,
  ({ one, many }) => ({
    plan: one(plans, {
      fields: [businessSubscriptions.planId],
      references: [plans.id],
    }),
    business: one(business, {
      fields: [businessSubscriptions.businessId],
      references: [business.id],
    }),
    payments: many(paymentRecords),
  }),
);

export type BusinessSubscription = typeof businessSubscriptions.$inferSelect;
export type NewBusinessSubscription = typeof businessSubscriptions.$inferInsert;
