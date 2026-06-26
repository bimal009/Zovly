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
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";
import { categories } from "./category";
import { productVariants } from "./product-variants";

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
