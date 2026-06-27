import {
  pgTable,
  text,
  timestamp,
  uuid,
  integer,
  numeric,
  jsonb,
  index,
  pgEnum,
  customType,
} from "drizzle-orm/pg-core";
import { relations, sql } from "drizzle-orm";
import { business } from "./business";
import { categories } from "./category";
import { productVariants } from "./product-variants";

// Postgres full-text search vector. Drizzle has no native tsvector type, so we
// declare a thin custom type. The column is GENERATED ALWAYS (STORED) from
// name + description + tags and indexed with GIN for `@@ websearch_to_tsquery`.
// NOTE: requires `CREATE EXTENSION IF NOT EXISTS pg_trgm;` to be run once on the
// database for the trigram index below (idx_products_name_trgm) to apply.
const tsvector = customType<{ data: string }>({
  dataType() {
    return "tsvector";
  },
});

// ─── Enums ───────────────────────────────────────────────────────────────────

export const productStatusEnum = pgEnum("product_status", [
  "active",
  "inactive",
  "archived",
]);

// ─── Products ────────────────────────────────────────────────────────────────

export const products = pgTable(
  "products",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    categoryId: uuid("category_id").references(() => categories.id, {
      onDelete: "set null",
    }),

    name: text("name").notNull(),
    description: text("description"), // fed into RAG / AI DM context
    sku: text("sku"),
    status: productStatusEnum("status").notNull().default("active"),

    tags: text("tags").array().default([]), // ["winter", "waterproof", "unisex"]

    // structured product-level attributes — { "material": "cotton", "fit": "slim" }
    // fed into RAG / AI DM context alongside description + tags
    attributes: jsonb("attributes"),

    price: numeric("price", { precision: 12, scale: 2 }).notNull(), // selling price — what customer pays
    costPrice: numeric("cost_price", { precision: 12, scale: 2 }), // what business paid (never shown to customer)
    discount: integer("discount").notNull().default(0), // percentage 0-100
    currency: text("currency").notNull().default("NPR"),

    stockQty: integer("stock_qty").notNull().default(0),
    lowStockThreshold: integer("low_stock_threshold").default(5),

    images: text("images").array().default([]),

    // Lexical search vector — fused with semantic + trigram in the hybrid
    // product-discovery search. Generated/stored so it stays in sync on write.
    // NOTE: a GENERATED column requires an IMMUTABLE expression. array_to_string()
    // is only STABLE, so instead of joining tags into the text we append them with
    // array_to_tsvector() (IMMUTABLE) to the name+description vector.
    searchTsv: tsvector("search_tsv").generatedAlwaysAs(
      sql`to_tsvector('simple', coalesce(name, '') || ' ' || coalesce(description, '')) || array_to_tsvector(array_remove(coalesce(tags, '{}'::text[]), NULL))`,
    ),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("products_business_id_idx").on(table.businessId),
    index("products_category_id_idx").on(table.categoryId),
    index("products_status_idx").on(table.status),
    index("products_sku_idx").on(table.sku),
    index("products_business_status_idx").on(table.businessId, table.status),
    // Full-text (lexical) search over name/description/tags.
    index("idx_products_search_tsv").using("gin", table.searchTsv),
    // Trigram index on name — catches typos and Romanized spellings via `%`.
    index("idx_products_name_trgm").using(
      "gin",
      sql`${table.name} gin_trgm_ops`,
    ),
  ],
);

export const productRelations = relations(products, ({ one, many }) => ({
  business: one(business, {
    fields: [products.businessId],
    references: [business.id],
  }),
  category: one(categories, {
    fields: [products.categoryId],
    references: [categories.id],
  }),
  variants: many(productVariants),
}));

export type Product = typeof products.$inferSelect;
export type NewProduct = typeof products.$inferInsert;
