import {
  pgTable,
  text,
  timestamp,
  uuid,
  boolean,
  index,
  unique,
} from "drizzle-orm/pg-core";
import { business } from "./business";
import { messages } from "./messages";
import { products } from "./products";
import { platformEnum } from "./enums";
import { relations } from "drizzle-orm";

export const conversations = pgTable(
  "conversations",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    platform: platformEnum("platform").notNull(),
    threadId: text("thread_id").notNull(), // platform-native thread/conversation ID
    contactId: text("contact_id").notNull(), // platform-native user/contact ID

    contactName: text("contact_name"),
    contactUsername: text("contact_username"), // IG @handle; null for Messenger
    contactAvatarUrl: text("contact_avatar_url"),

    lastMessageAt: timestamp("last_message_at").defaultNow().notNull(),

    // the product currently under discussion (for follow-up context)
    activeProductId: uuid("active_product_id").references(() => products.id, {
      onDelete: "set null",
    }),
    activeProductAt: timestamp("active_product_at"), // when it was set (for staleness)

    createdAt: timestamp("created_at").defaultNow().notNull(),
  },
  (table) => [
    // fast lookup by business + platform (inbox view)
    index("conv_business_platform_idx").on(table.businessId, table.platform),
    // uniqueness: one thread per contact per platform per business
    unique("conv_thread_uq").on(
      table.businessId,
      table.platform,
      table.threadId,
    ),
  ],
);

export const conversationsRelations = relations(
  conversations,
  ({ one, many }) => ({
    business: one(business, {
      fields: [conversations.businessId],
      references: [business.id],
    }),
    activeProduct: one(products, {
      fields: [conversations.activeProductId],
      references: [products.id],
    }),
    messages: many(messages),
  }),
);