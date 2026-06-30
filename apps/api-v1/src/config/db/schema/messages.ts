import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  uuid,
  boolean,
  index,
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

    direction: messageDirectionEnum("direction").notNull(), // in | out
    sentBy: messageSenderEnum("sent_by"), // ai | human — null for inbound

    // Content — text or voice transcript
    content: text("content"),

    // Media — ImageKit signed URL + type (image/video/audio/document)
    mediaUrl: text("media_url"),
    mediaType: messageMediaTypeEnum("media_type"),

    // Vectorized flag — py-ml sets this true after embedding
    isVectorized: boolean("is_vectorized").default(false).notNull(),
    // messages table — add:
    status: messageStatusEnum("status"), // null for inbound; set for outbound
    errorMessage: text("error_message"), // failure reason when status = failed
    sentToPlatformAt: timestamp("sent_to_platform_at"), // when Graph API accepted
    platformMessageId: text("platform_message_id"), // Meta's returned message id

    sentAt: timestamp("sent_at").defaultNow().notNull(),
  },
  (table) => [
    // pull all messages for a conversation ordered by time
    index("msg_conversation_idx").on(table.conversationId),
    // filter un-vectorized messages for the embedding worker
    index("msg_vectorize_idx").on(table.businessId, table.isVectorized),
  ],
);

export type Conversation = typeof conversations.$inferSelect;
export type NewConversation = typeof conversations.$inferInsert;
export type Message = typeof messages.$inferSelect;
export type NewMessage = typeof messages.$inferInsert;

export const messagesRelations = relations(messages, ({ one }) => ({
  conversation: one(conversations, { fields: [messages.conversationId], references: [conversations.id] }),
  business: one(business, { fields: [messages.businessId], references: [business.id] }),
}));
