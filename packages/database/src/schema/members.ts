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

export const inviteStatusEnum = pgEnum("invite_status", [
  "pending", // email sent, not accepted yet
  "accepted", // user signed up and joined
  "declined", // user declined
  "expired", // 7-day window passed
  "revoked", // owner/admin cancelled before acceptance
]);

export const permissionActionEnum = pgEnum("permission_action", [
  "read",
  "write",
  "delete",
  "manage", // read + write + delete
]);

// ── Business Members ──────────────────────────────────────────────────────────
// Every user that belongs to a business lives here.

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

    // ── Per-module permission overrides ───────────────────────────────────────
    // These override the role defaults for this specific member.
    // null = inherit from role, true/false = explicit override

    // Content
    canManageContent: boolean("can_manage_content"), // posts, scheduling
    canViewAnalytics: boolean("can_view_analytics"),
    canManageAds: boolean("can_manage_ads"),

    // Inbox
    canReadDms: boolean("can_read_dms"),
    canReplyDms: boolean("can_reply_dms"),
    canReadComments: boolean("can_read_comments"),
    canReplyComments: boolean("can_reply_comments"),

    // Leads
    canViewLeads: boolean("can_view_leads"),
    canManageLeads: boolean("can_manage_leads"),

    // Bookings
    canViewBookings: boolean("can_view_bookings"),
    canManageBookings: boolean("can_manage_bookings"),

    // Inventory & Orders
    canViewInventory: boolean("can_view_inventory"),
    canManageInventory: boolean("can_manage_inventory"),
    canViewOrders: boolean("can_view_orders"),

    // Settings
    canManageSettings: boolean("can_manage_settings"), // app connections, AI persona
    canManageMembers: boolean("can_manage_members"), // invite/remove users
    canManageBilling: boolean("can_manage_billing"), // owner only normally

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

// ── Member Invites ────────────────────────────────────────────────────────────
// Invite flow: owner/admin creates invite → email sent → user accepts via token

export const memberInvites = pgTable(
  "member_invites",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull(),

    // Who sent the invite
    invitedById: text("invited_by_id").references(() => user.id, {
      onDelete: "set null",
    }),

    // Who is being invited
    invitedEmail: text("invited_email").notNull(),

    // Role they'll get on acceptance
    role: memberRoleEnum("role").notNull().default("viewer"),

    // ── Same permission overrides as businessMembers ───────────────────────
    // Set at invite time, copied to businessMembers on acceptance
    canManageContent: boolean("can_manage_content"),
    canViewAnalytics: boolean("can_view_analytics"),
    canManageAds: boolean("can_manage_ads"),
    canReadDms: boolean("can_read_dms"),
    canReplyDms: boolean("can_reply_dms"),
    canReadComments: boolean("can_read_comments"),
    canReplyComments: boolean("can_reply_comments"),
    canViewLeads: boolean("can_view_leads"),
    canManageLeads: boolean("can_manage_leads"),
    canViewBookings: boolean("can_view_bookings"),
    canManageBookings: boolean("can_manage_bookings"),
    canViewInventory: boolean("can_view_inventory"),
    canManageInventory: boolean("can_manage_inventory"),
    canViewOrders: boolean("can_view_orders"),
    canManageSettings: boolean("can_manage_settings"),
    canManageMembers: boolean("can_manage_members"),
    canManageBilling: boolean("can_manage_billing"),

    // ── Invite token ──────────────────────────────────────────────────────────
    token: text("token").notNull().unique(), // crypto.randomUUID() — in email link
    status: inviteStatusEnum("status").notNull().default("pending"),
    expiresAt: timestamp("expires_at").notNull(), // defaultNow() + 7 days

    acceptedAt: timestamp("accepted_at"),
    declinedAt: timestamp("declined_at"),
    revokedAt: timestamp("revoked_at"),
    revokedById: text("revoked_by_id").references(() => user.id, {
      onDelete: "set null",
    }),

    createdAt: timestamp("created_at").defaultNow().notNull(),
    updatedAt: timestamp("updated_at")
      .defaultNow()
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [
    index("invite_business_idx").on(table.businessId),
    index("invite_email_idx").on(table.invitedEmail),
    index("invite_token_idx").on(table.token),
    index("invite_status_idx").on(table.status),
  ],
);

// ── Type helpers ──────────────────────────────────────────────────────────────

export type BusinessMember = typeof businessMembers.$inferSelect;
export type NewBusinessMember = typeof businessMembers.$inferInsert;
export type MemberInvite = typeof memberInvites.$inferSelect;
export type NewMemberInvite = typeof memberInvites.$inferInsert;
export type MemberRole = (typeof memberRoleEnum.enumValues)[number];
