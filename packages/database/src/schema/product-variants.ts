import {
  pgTable,
  text,
  timestamp,
  uuid,
  integer,
  jsonb,
  index,
  unique,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";
import { products } from "./products";

export const productVariants = pgTable(
  "product_variants",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    productId: uuid("product_id")
      .notNull()
      .references(() => products.id, { onDelete: "cascade" }),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    name: text("name").notNull(), // "Red / Medium", "500ml", "Large"
    sku: text("sku"),

    attributes: jsonb("attributes"),

    price: integer("price"), // selling price override; null → use product.price
    costPrice: integer("cost_price"), // what business paid (never shown to customer)
    discount: integer("discount"), // override; null → use product.discount

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
    index("pv_product_idx").on(table.productId),
    index("pv_business_idx").on(table.businessId),
    index("pv_sku_idx").on(table.sku),
    // no two variants of the same product share a name ("Red / M" once per product)
    unique("pv_product_name_uq").on(table.productId, table.name),
  ],
);

export const productVariantRelations = relations(
  productVariants,
  ({ one }) => ({
    product: one(products, {
      fields: [productVariants.productId],
      references: [products.id],
    }),
    business: one(business, {
      fields: [productVariants.businessId],
      references: [business.id],
    }),
  }),
);

export type ProductVariant = typeof productVariants.$inferSelect;
export type NewProductVariant = typeof productVariants.$inferInsert;
