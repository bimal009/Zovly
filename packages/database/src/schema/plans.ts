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
import { businessSubscriptions } from "./subscriptions";
import { paymentRecords } from "./payments";

export const plans = pgTable(
  "plans",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    name: text("name").notNull().unique(),

    paddleProductId: text("paddle_product_id").unique(),
    paddlePriceIdMonthly: text("paddle_price_id_monthly").unique(),
    paddlePriceIdYearly: text("paddle_price_id_yearly").unique(),

    monthlyPrice: integer("monthly_price").notNull(),
    yearlyPrice: integer("yearly_price").notNull(),

    maxMembers: integer("max_members").notNull(),
    maxSocialAccounts: integer("max_social_accounts").notNull(),
    maxAiRepliesMonth: integer("max_ai_replies_month").notNull(),
    maxPostsMonth: integer("max_posts_month").notNull(),
    maxLeads: integer("max_leads").notNull(),
    maxProducts: integer("max_products").notNull(),
    maxBookingsMonth: integer("max_bookings_month").notNull(),

    hasVideoUpload: boolean("has_video_upload").default(false).notNull(),
    hasMultiPlatformPost: boolean("has_multi_platform_post")
      .default(false)
      .notNull(),
    hasPostAnalytics: boolean("has_post_analytics").default(false).notNull(),
    hasAiDmReplies: boolean("has_ai_dm_replies").default(false).notNull(),
    hasAiCommentReplies: boolean("has_ai_comment_replies")
      .default(false)
      .notNull(),
    hasAiLeadScoring: boolean("has_ai_lead_scoring").default(false).notNull(),
    hasAiAdSuggestions: boolean("has_ai_ad_suggestions")
      .default(false)
      .notNull(),
    hasVoiceTranscription: boolean("has_voice_transcription")
      .default(false)
      .notNull(),
    hasImageUnderstanding: boolean("has_image_understanding")
      .default(false)
      .notNull(),
    hasBookings: boolean("has_bookings").default(false).notNull(),
    hasInventory: boolean("has_inventory").default(false).notNull(),
    hasPayments: boolean("has_payments").default(false).notNull(),
    hasGoogleWorkspace: boolean("has_google_workspace")
      .default(false)
      .notNull(),
    hasMetaAds: boolean("has_meta_ads").default(false).notNull(),
    hasTikTokAds: boolean("has_tiktok_ads").default(false).notNull(),
    hasPrioritySupport: boolean("has_priority_support")
      .default(false)
      .notNull(),

    aiReplyOveragePriceUsdPer500: integer("ai_reply_overage_price_usd_per_500"),
    isActive: boolean("is_active").default(true).notNull(),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("plan_name_idx").on(table.name),
    index("plan_active_idx").on(table.isActive),
    index("plan_paddle_product_idx").on(table.paddleProductId),
    index("plan_paddle_monthly_idx").on(table.paddlePriceIdMonthly),
    index("plan_paddle_yearly_idx").on(table.paddlePriceIdYearly),
  ],
);

export const planRelations = relations(plans, ({ many }) => ({
  subscriptions: many(businessSubscriptions),
  payments: many(paymentRecords),
}));

export type Plan = typeof plans.$inferSelect;
export type NewPlan = typeof plans.$inferInsert;
