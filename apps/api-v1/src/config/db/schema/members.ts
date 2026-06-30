import { relations } from "drizzle-orm";
import { business } from "./business";
import { user } from "./user";
import {
  pgTable,
  pgEnum,
  uuid,
  text,
  boolean,
  timestamp,
  index,
  unique,
} from "drizzle-orm/pg-core";

// ── Enums ─────────────────────────────────────────────────────────────────────

export const memberRoleEnum = pgEnum("member_role", [
  "owner", // full access, billing, delete business — assigned on signup
  "admin", // full access except billing/delete
  "manager", // can manage content, leads, bookings — no settings
  "staff", // reply to DMs, view bookings only
  "viewer", // read-only across everything
]);

export const permissionActionEnum = pgEnum("permission_action", [
  "read",
  "write",
  "delete",
  "manage", // read + write + delete
]);

export const businessMembers = pgTable(
  "business_members",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull(),

    userId: text("user_id")
      .references(() => user.id, { onDelete: "cascade" })
      .notNull(),

    role: memberRoleEnum("role").notNull().default("viewer"),

    canManageContent: boolean("can_manage_content").notNull().default(false), // posts, scheduling
    canViewAnalytics: boolean("can_view_analytics").notNull().default(false),
    canManageAds: boolean("can_manage_ads").notNull().default(false),

    canReadDms: boolean("can_read_dms").notNull().default(false),
    canReplyDms: boolean("can_reply_dms").notNull().default(false),
    canReadComments: boolean("can_read_comments").notNull().default(false),
    canReplyComments: boolean("can_reply_comments").notNull().default(false),

    canViewLeads: boolean("can_view_leads").notNull().default(false),
    canManageLeads: boolean("can_manage_leads").notNull().default(false),

    canViewBookings: boolean("can_view_bookings").notNull().default(false),
    canManageBookings: boolean("can_manage_bookings").notNull().default(false),

    canViewInventory: boolean("can_view_inventory").notNull().default(false),
    canManageInventory: boolean("can_manage_inventory")
      .notNull()
      .default(false),
    canViewOrders: boolean("can_view_orders").notNull().default(false),

    // Settings
    canManageSettings: boolean("can_manage_settings").notNull().default(false), // app connections, AI persona
    canManageMembers: boolean("can_manage_members").notNull().default(false), // invite/remove users
    canManageBilling: boolean("can_manage_billing").notNull().default(false), // owner only normally

    // ── Lifecycle ─────────────────────────────────────────────────────────────
    joinedAt: timestamp("joined_at").defaultNow().notNull(),
    lastSeenAt: timestamp("last_seen_at"),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    // One membership per user per business
    unique("uq_business_member").on(table.businessId, table.userId),
    index("biz_member_business_idx").on(table.businessId),
    index("biz_member_user_idx").on(table.userId),
    index("biz_member_role_idx").on(table.role),
  ],
);

export type BusinessMember = typeof businessMembers.$inferSelect;
export type NewBusinessMember = typeof businessMembers.$inferInsert;
export type MemberRole = (typeof memberRoleEnum.enumValues)[number];

export const businessMembersRelations = relations(businessMembers, ({ one }) => ({
  business: one(business, { fields: [businessMembers.businessId], references: [business.id] }),
  user: one(user, { fields: [businessMembers.userId], references: [user.id] }),
}));
