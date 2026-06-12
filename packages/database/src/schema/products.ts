// db/schema/products.ts
import {
  pgTable,
  text,
  timestamp,
  uuid,
  integer,
  index,
  pgEnum,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";

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

    // ── Core ──────────────────────────────────────────────────────────────
    name: text("name").notNull(),
    description: text("description"), // fed into RAG / AI DM context
    sku: text("sku"),
    status: productStatusEnum("status").notNull().default("active"),

    // ── Pricing (all in cents) ────────────────────────────────────────────
    price: integer("price").notNull(), // selling price — what customer pays
    costPrice: integer("cost_price"), // what business paid/made it for (never shown to customer)
    discount: integer("discount").notNull().default(0), // percentage 0-100
    currency: text("currency").notNull().default("NPR"),

    // ── Inventory ─────────────────────────────────────────────────────────
    stockQty: integer("stock_qty").notNull().default(0),
    lowStockThreshold: integer("low_stock_threshold").default(5),

    // ── Media (ImageKit URLs) ─────────────────────────────────────────────
    images: text("images").array().default([]),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("products_business_id_idx").on(table.businessId),
    index("products_status_idx").on(table.status),
    index("products_sku_idx").on(table.sku),
    index("products_business_status_idx").on(table.businessId, table.status),
  ],
);

export const productRelations = relations(products, ({ one }) => ({
  business: one(business, {
    fields: [products.businessId],
    references: [business.id],
  }),
}));

export type Product = typeof products.$inferSelect;
export type NewProduct = typeof products.$inferInsert;
