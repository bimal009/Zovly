import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  uuid,
  boolean,
  index,
  type AnyPgColumn,
} from "drizzle-orm/pg-core";
import { business } from "./business";
import { conversations } from "./conversations";
import {
  platformEnum,
  messageDirectionEnum,
  messageSenderEnum,
  messageMediaTypeEnum,
  messageStatusEnum,
} from "./enums";

export {
  platformEnum,
  messageDirectionEnum,
  messageSenderEnum,
  messageMediaTypeEnum,
};

export const messages = pgTable(
  "messages",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    conversationId: uuid("conversation_id")
      .notNull()
      .references(() => conversations.id, { onDelete: "cascade" }),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    direction: messageDirectionEnum("direction").notNull(),
    sentBy: messageSenderEnum("sent_by"),

    content: text("content"),

    mediaUrl: text("media_url"),
    mediaType: messageMediaTypeEnum("media_type"),


    status: messageStatusEnum("status"), 
    errorMessage: text("error_message"), 
    sentToPlatformAt: timestamp("sent_to_platform_at"), // when Graph API accepted
    platformMessageId: text("platform_message_id"), // Meta's returned message id (mid)

    deliveredAt: timestamp("delivered_at"), 
    seenAt: timestamp("seen_at"), 

    replyToMessageId: uuid("reply_to_message_id").references(
      (): AnyPgColumn => messages.id,
      { onDelete: "set null" }
    ),

    platformSenderId: text("platform_sender_id"),

    sentAt: timestamp("sent_at").defaultNow().notNull(),
  },
  (table) => [
    index("msg_conversation_idx").on(table.conversationId),
    index("msg_platform_id_idx").on(table.platformMessageId),
    index("msg_reply_to_idx").on(table.replyToMessageId),
  ],
);

export type Conversation = typeof conversations.$inferSelect;
export type NewConversation = typeof conversations.$inferInsert;
export type Message = typeof messages.$inferSelect;
export type NewMessage = typeof messages.$inferInsert;

export const messagesRelations = relations(messages, ({ one }) => ({
  conversation: one(conversations, { fields: [messages.conversationId], references: [conversations.id] }),
  business: one(business, { fields: [messages.businessId], references: [business.id] }),
  replyTo: one(messages, {
    fields: [messages.replyToMessageId],
    references: [messages.id],
    relationName: "replyTo",
  }),
}));