import {
  boolean,
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
  "product",
  "service",
]);

export const knowledgeChunks = pgTable(
  "knowledge_chunks",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),
    sourceType: knowledgeSourceTypeEnum("source_type").notNull(),
    sourceId: uuid("source_id").notNull(), // id of the faq/policy/post/product/service row — polymorphic, NO FK
    chunkIndex: integer("chunk_index").default(0).notNull(),
    content: text("content").notNull(), // clean chunk text (no e5 prefix)
    embedding: vector("embedding", { dimensions: 1024 }).notNull(),
    metadata: jsonb("metadata"), // faq: { question }; post: { url, publishedAt }; product: { price, currency, sku, inStock }
    isActive: boolean("is_active").default(true).notNull(), // mirrors source row's active state; excluded from retrieval when false
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