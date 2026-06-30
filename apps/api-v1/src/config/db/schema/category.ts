// db/schema/categories.ts
import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  uuid,
  index,
  unique,
} from "drizzle-orm/pg-core";
import { business } from "./business";
import { products } from "./products";

export const categories = pgTable(
  "categories",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    name: text("name").notNull(), // "Jackets", "Footwear"
    description: text("description"), // optional, can feed RAG context
    slug: text("slug"), // url-friendly, optional

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("categories_business_idx").on(table.businessId),
    // a business can't have two categories with the same name
    unique("categories_business_name_uq").on(table.businessId, table.name),
  ],
);

export type Category = typeof categories.$inferSelect;
export type NewCategory = typeof categories.$inferInsert;

export const categoriesRelations = relations(categories, ({ one, many }) => ({
  business: one(business, { fields: [categories.businessId], references: [business.id] }),
  products: many(products),
}));
