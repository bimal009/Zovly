import {
  index,
  integer,
  jsonb,
  pgEnum,
  pgTable,
  text,
  timestamp,
  unique,
  uuid,
  vector,
} from "drizzle-orm/pg-core";
import { business } from "./business";

export const knowledgeSourceTypeEnum = pgEnum("knowledge_source_type", [
  "faq",
  "policy",
  "post",
]);

export const knowledgeChunks = pgTable(
  "knowledge_chunks",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),
    sourceType: knowledgeSourceTypeEnum("source_type").notNull(),
    sourceId: uuid("source_id").notNull(), // id of the faq/policy/post row — polymorphic, NO FK
    chunkIndex: integer("chunk_index").default(0).notNull(),
    content: text("content").notNull(), // clean chunk text (no e5 prefix)
    embedding: vector("embedding", { dimensions: 1024 }).notNull(),
    tokenCount: integer("token_count"), // precomputed for the RAG token-budget fill
    metadata: jsonb("metadata"), // faq: { question }; post: { url, publishedAt }
    createdAt: timestamp("created_at").defaultNow().notNull(),
  },
  (table) => [
    index("kc_hnsw_idx").using("hnsw", table.embedding.op("vector_cosine_ops")),
    index("kc_business_idx").on(table.businessId),
    index("kc_source_idx").on(
      table.businessId,
      table.sourceType,
      table.sourceId,
    ),
    unique("kc_source_chunk_uq").on(
      table.sourceType,
      table.sourceId,
      table.chunkIndex,
    ),
  ],
);
