// db/schema/business.ts
import {
  pgEnum,
  pgTable,
  text,
  timestamp,
  uuid,
  index,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { businessMembers } from "./members";
import { appConnections } from "./apps";
import { products } from "./products";
import { services } from "./services";

export const businessTypeEnum = pgEnum("business_type", [
  "product",
  "service",
  "both",
]);

export const planEnum = pgEnum("plan", ["starter", "growth", "pro", "agency"]);

export const business = pgTable(
  "business",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    name: text("name").notNull(),
    description: text("description"),
    logo: text("logo"), // R2 key
    website: text("website"),
    phone: text("phone"),
    address: text("address"),
    city: text("city"),
    country: text("country").default("NPL"), // Nepal default

    type: businessTypeEnum("type").notNull().default("service"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("business_type_idx").on(table.type)],
);

export const businessRelations = relations(business, ({ many, one }) => ({
  members: many(businessMembers),
  appConnections: one(appConnections, {
    fields: [business.id],
    references: [appConnections.businessId],
  }),
  products: many(products),
  services: many(services),
}));
export type Business = typeof business.$inferSelect;
export type NewBusiness = typeof business.$inferInsert;
