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

    appName: text("app_name").notNull(),

    accessToken: text("access_token"), // encrypted
    refreshToken: text("refresh_token"), // encrypted
    tokenExpiresAt: timestamp("token_expires_at"),
    scopes: text("scopes").array(),

    publicKey: text("public_key"), // encrypted
    secretKey: text("secret_key"), // encrypted
    merchantId: text("merchant_id"),

    platformAccountId: text("platform_account_id"), // IG user_id, WA phone_number_id …
    platformAccountName: text("platform_account_name"), // shown in UI

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
    // One credential per app per business (one-to-one): a business can connect
    // at most one facebook, one instagram, one whatsapp, ... account.
    unique("app_cred_app_uq").on(table.businessId, table.appName),
  ],
);
export type AppCredential = typeof appCredentials.$inferSelect;
export type NewAppCredential = typeof appCredentials.$inferInsert;
