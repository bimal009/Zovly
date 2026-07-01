import { relations } from "drizzle-orm";
import { business } from "./business";
import {
  pgTable,
  uuid,
  text,
  boolean,
  timestamp,
  index,
  unique,
} from "drizzle-orm/pg-core";

export const appCredentials = pgTable(
  "app_credentials",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull(),

    appName: text("app_name").notNull(), // "facebook" | "instagram" | "whatsapp" | ...

    accessToken: text("access_token"), // encrypted
    refreshToken: text("refresh_token"), // encrypted
    tokenExpiresAt: timestamp("token_expires_at"),
    scopes: text("scopes").array(),

    publicKey: text("public_key"), // encrypted
    secretKey: text("secret_key"), // encrypted
    merchantId: text("merchant_id"),

    // Identifies the specific external account/page — required now that a
    // business can connect MULTIPLE accounts for the same app.
    platformAccountId: text("platform_account_id").notNull(), // FB Page id, IG user_id, WA phone_number_id …
    platformAccountName: text("platform_account_name"), // shown in UI
    platformAccountImage: text("platform_account_image"), // Page/profile picture URL, shown in UI

    webhookVerifyToken: text("webhook_verify_token"),
    webhookSubscribedAt: timestamp("webhook_subscribed_at"),
    isActive: boolean("is_active").default(true).notNull(),
    connectedAt: timestamp("connected_at"),
    disconnectedAt: timestamp("disconnected_at"),
    lastSyncAt: timestamp("last_sync_at"),
    errorMessage: text("error_message"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("app_cred_business_idx").on(table.businessId),
    index("app_cred_app_name_idx").on(table.appName),
    index("app_cred_active_idx").on(table.businessId, table.appName, table.isActive),
    // A business can connect multiple accounts of the same app (e.g. two
    // Facebook Pages), but the SAME external account can't be linked twice.
    unique("app_cred_account_uq").on(table.businessId, table.appName, table.platformAccountId),
  ],
);

export type AppCredential = typeof appCredentials.$inferSelect;
export type NewAppCredential = typeof appCredentials.$inferInsert;

export const appCredentialsRelations = relations(appCredentials, ({ one }) => ({
  business: one(business, { fields: [appCredentials.businessId], references: [business.id] }),
}));