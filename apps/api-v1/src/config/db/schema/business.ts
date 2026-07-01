// db/schema/business.ts
import { relations } from "drizzle-orm";
import {
  pgEnum,
  pgTable,
  text,
  timestamp,
  uuid,
  index,
} from "drizzle-orm/pg-core";
import { businessMembers } from "./members";
import { products } from "./products";
import { services } from "./services";
import { policies } from "./policies";

export const businessTypeEnum = pgEnum("business_type", [
  "product",
  "service",
  "both",
]);

export const planEnum = pgEnum("plan", ["starter", "growth", "pro"]);

export const business = pgTable(
  "business",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    name: text("name").notNull(),
    description: text("description").notNull(),
    logo: text("logo"), //link to imagekit
    website: text("website"),
    phone: text("phone"),
    address: text("address").notNull(),
    city: text("city"),
    country: text("country").default("NPL"),

    type: businessTypeEnum("type").notNull().default("service"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("business_type_idx").on(table.type)],
);

export type Business = typeof business.$inferSelect;
export type NewBusiness = typeof business.$inferInsert;

export const businessRelations = relations(business, ({ one, many }) => ({
  members: many(businessMembers),

  products: many(products),
  services: many(services),
  policies: many(policies),
}));
