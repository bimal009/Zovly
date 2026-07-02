import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  uuid,
  boolean,
  index,
  unique,
  uniqueIndex,
} from "drizzle-orm/pg-core";
import { business } from "./business";
import { messages } from "./messages";
import { platformEnum } from "./enums";

export const conversations = pgTable(
  "conversations",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    platform: platformEnum("platform").notNull(),
    threadId: text("thread_id").notNull(),
    contactId: text("contact_id").notNull(), 

    contactName: text("contact_name"),
    contactUsername: text("contact_username"), 
    contactAvatarUrl: text("contact_avatar_url"),

    lastMessageAt: timestamp("last_message_at").defaultNow().notNull(),



    createdAt: timestamp("created_at").defaultNow().notNull(),
  },
  (table) => [
    index("conv_business_platform_idx").on(table.businessId, table.platform),
    uniqueIndex("convo_business_thread_idx").on(table.businessId, table.threadId),
    unique("conv_thread_uq").on(
      table.businessId,
      table.platform,
      table.threadId,
    ),
  ],
);

export const conversationsRelations = relations(conversations, ({ one, many }) => ({
  business: one(business, { fields: [conversations.businessId], references: [business.id] }),
  messages: many(messages),
}));
