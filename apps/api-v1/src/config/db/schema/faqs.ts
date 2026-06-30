// db/schema/faqs.ts
import { relations } from "drizzle-orm";
import {
  pgTable,
  text,
  timestamp,
  uuid,
  index,
  boolean,
  integer,
} from "drizzle-orm/pg-core";
import { business } from "./business";

export const faqs = pgTable(
  "faqs",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),
    question: text("question").notNull(),
    answer: text("answer").notNull(),
    isActive: boolean("is_active").default(true).notNull(), // toggle off without deleting
    sortOrder: integer("sort_order").default(0).notNull(), // display order in the UI
    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("faq_business_idx").on(table.businessId)],
);

export type Faq = typeof faqs.$inferSelect;
export type NewFaq = typeof faqs.$inferInsert;

export const faqsRelations = relations(faqs, ({ one }) => ({
  business: one(business, { fields: [faqs.businessId], references: [business.id] }),
}));
