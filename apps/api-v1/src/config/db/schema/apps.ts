import { relations } from "drizzle-orm";
import { business } from "./business";
import {
  pgTable,
  uuid,
  text,
  boolean,
  timestamp,
  index,
} from "drizzle-orm/pg-core";

export const appConnections = pgTable(
  "app_connections",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull()
      .unique(), // one row per business, always

    instagram: boolean("instagram").default(false).notNull(),
    facebook: boolean("facebook").default(false).notNull(),
    tiktok: boolean("tiktok").default(false).notNull(),
    whatsapp: boolean("whatsapp").default(false).notNull(),

    googleWorkspace: boolean("google_workspace").default(false).notNull(),

    stripeConnect: boolean("stripe_connect").default(false).notNull(),

    fonepay: boolean("fonepay").default(false).notNull(),
    khalti: boolean("khalti").default(false).notNull(),
    esewa: boolean("esewa").default(false).notNull(),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("app_conn_business_idx").on(table.businessId)],
);

export type AppConnections = typeof appConnections.$inferSelect;
export type NewAppConnections = typeof appConnections.$inferInsert;

export const appConnectionsRelations = relations(appConnections, ({ one }) => ({
  business: one(business, { fields: [appConnections.businessId], references: [business.id] }),
}));
