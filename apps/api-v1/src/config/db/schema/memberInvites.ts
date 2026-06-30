import { business } from "./business";
import { memberRoleEnum } from "./members";
import { user } from "./user";
import {
  pgTable,
  pgEnum,
  uuid,
  text,
  timestamp,
  index,
} from "drizzle-orm/pg-core";

export const inviteStatusEnum = pgEnum("invite_status", [
  "pending", // email sent, not accepted yet
  "accepted", // user signed up and joined
  "declined", // user declined
  "expired", // 7-day window passed
  "revoked", // owner/admin cancelled before acceptance
]);

export const memberInvites = pgTable(
  "member_invites",
  {
    id: uuid("id").primaryKey().defaultRandom(),

    businessId: uuid("business_id")
      .references(() => business.id, { onDelete: "cascade" })
      .notNull(),

    invitedById: text("invited_by_id").references(() => user.id, {
      onDelete: "set null",
    }),

    invitedEmail: text("invited_email").notNull(),

    role: memberRoleEnum("role").notNull().default("viewer"),

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
export type MemberInvite = typeof memberInvites.$inferSelect;
export type NewMemberInvite = typeof memberInvites.$inferInsert;
