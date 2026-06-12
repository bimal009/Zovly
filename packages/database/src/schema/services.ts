// db/schema/services.ts
import {
  pgTable,
  text,
  timestamp,
  uuid,
  integer,
  boolean,
  index,
  pgEnum,
  jsonb,
} from "drizzle-orm/pg-core";
import { relations } from "drizzle-orm";
import { business } from "./business";

// ─── Enums ───────────────────────────────────────────────────────────────────

export const serviceTypeEnum = pgEnum("service_type", [
  "appointment", // haircut, consultation — has a booking slot
  "membership", // gym monthly, app sub — recurring billing
  "class", // yoga, workshop — slot + multi-person
  "package", // "10 sessions" — one-time, redeemed later
]);

export const serviceStatusEnum = pgEnum("service_status", [
  "active",
  "inactive",
  "archived",
]);

export const billingIntervalEnum = pgEnum("billing_interval", [
  "weekly",
  "monthly",
  "quarterly",
  "yearly",
]);

// ─── Services ────────────────────────────────────────────────────────────────

export const services = pgTable(
  "services",
  {
    id: uuid("id").primaryKey().defaultRandom(),
    businessId: uuid("business_id")
      .notNull()
      .references(() => business.id, { onDelete: "cascade" }),

    // ── Core (all types) ──────────────────────────────────────────────────
    type: serviceTypeEnum("type").notNull().default("appointment"),
    status: serviceStatusEnum("status").notNull().default("active"),
    name: text("name").notNull(),
    description: text("description"), // fed into RAG / AI DM context

    // ── Pricing (all in cents) ────────────────────────────────────────────
    price: integer("price").notNull(), // what customer pays
    costPrice: integer("cost_price"), // business cost / overhead (never shown)
    mrp: integer("mrp"), // shown crossed out in UI + AI captions
    currency: text("currency").notNull().default("NPR"),

    // ── Payment handling ──────────────────────────────────────────────────
    // Stripe connected  → session built dynamically from price + name + currency
    // Not connected     → AI sends follow-up DM with price + contact info
    requiresDeposit: boolean("requires_deposit").notNull().default(false),
    depositAmount: integer("deposit_amount"), // cents, null = full upfront

    // ── appointment + class only ──────────────────────────────────────────
    durationMin: integer("duration_min"),
    bufferMin: integer("buffer_min").default(0),
    maxAdvanceDays: integer("max_advance_days").default(30),
    googleCalendarId: text("google_calendar_id"),

    // ── class only ────────────────────────────────────────────────────────
    maxConcurrent: integer("max_concurrent").default(1),

    // ── membership only ───────────────────────────────────────────────────
    // Stripe connected  → recurring subscription session created on the fly
    // Not connected     → AI follow-up with payment instructions
    billingInterval: billingIntervalEnum("billing_interval"),
    trialDays: integer("trial_days").default(0),

    // ── package only ──────────────────────────────────────────────────────
    sessionCount: integer("session_count"),
    validityDays: integer("validity_days"),

    // ── What's included (label/value pairs shown to customers) ───────────
    features: jsonb("features").$type<{ label: string; value: string }[]>().notNull().default([]),

    // ── Media (ImageKit URLs) ─────────────────────────────────────────────
    images: text("images").array().default([]),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("services_business_id_idx").on(table.businessId),
    index("services_type_idx").on(table.type),
    index("services_status_idx").on(table.status),
    index("services_business_status_idx").on(table.businessId, table.status),
  ],
);

export const serviceRelations = relations(services, ({ one }) => ({
  business: one(business, {
    fields: [services.businessId],
    references: [business.id],
  }),
}));

export type Service = typeof services.$inferSelect;
export type NewService = typeof services.$inferInsert;

// ─── Type narrowing helpers ───────────────────────────────────────────────────

export type AppointmentService = Service & { type: "appointment" };
export type MembershipService = Service & { type: "membership" };
export type ClassService = Service & { type: "class" };
export type PackageService = Service & { type: "package" };
