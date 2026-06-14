import {
  pgEnum,
  pgTable,
  text,
  timestamp,
  uuid,
  boolean,
  index,
} from "drizzle-orm/pg-core";
import { business } from "./business";
import { messages } from "./messages";
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

    aiEnabled: boolean("ai_enabled").default(true).notNull(),
    lastMessageAt: timestamp("last_message_at").defaultNow().notNull(),

    createdAt: timestamp("created_at").defaultNow().notNull(),
  },
  (table) => [
    // fast lookup by business + platform (inbox view)
    index("conv_business_platform_idx").on(table.businessId, table.platform),
    // uniqueness: one thread per contact per platform per business
    index("conv_thread_idx").on(
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
    messages: many(messages),
  }),
);
