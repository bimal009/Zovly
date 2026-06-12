import {
  pgTable,
  text,
  timestamp,
  uuid,
  index,
  vector,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { messages } from "./messages";
import { business } from "./business";
import { conversations } from "./conversations";

export const messageEmbeddings = pgTable(
  "message_embeddings",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    messageId: uuid("message_id")
      .notNull()
      .unique()
      .references(() => messages.id, { onDelete: "cascade" }),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),
    conversationId: uuid("conversation_id")
      .notNull()
      .references(() => conversations.id, { onDelete: "cascade" }),

    // text that was embedded (message content or voice transcript)
    content: text("content").notNull(),

    // nomic-embed-text-v2-moe — 768 dims
    embedding: vector("embedding", { dimensions: 1024 }).notNull(),

    createdAt: timestamp("created_at").defaultNow().notNull(),
  },
  (table) => [
    // HNSW cosine index — fast similarity search
    index("msg_emb_hnsw_idx").using(
      "hnsw",
      table.embedding.op("vector_cosine_ops"),
    ),
    index("msg_emb_business_idx").on(table.businessId),
    index("msg_emb_conversation_idx").on(table.conversationId),
  ],
);

export const messageEmbeddingsRelations = relations(
  messageEmbeddings,
  ({ one }) => ({
    message: one(messages, {
      fields: [messageEmbeddings.messageId],
      references: [messages.id],
    }),
    business: one(business, {
      fields: [messageEmbeddings.businessId],
      references: [business.id],
    }),
    conversation: one(conversations, {
      fields: [messageEmbeddings.conversationId],
      references: [conversations.id],
    }),
  }),
);

export type MessageEmbedding = typeof messageEmbeddings.$inferSelect;
export type NewMessageEmbedding = typeof messageEmbeddings.$inferInsert;
