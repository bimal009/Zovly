import { eq } from "drizzle-orm";
import { CreateBusinessInput } from "@repo/types";

import { db } from "../../config/db/db";
import { business } from "../../config/db/schema/business";
import { user } from "../../config/db/schema/user";
import { User } from "../../lib/auth";
import { businessMembers } from "../../config/db/schema/members";

export const create = async (session: User, input: CreateBusinessInput) => {
  if (!session) {
    throw new Error("User not found");
  }

  if (session.isOnboarded) {
    throw new Error("User already has a business");
  }

  return await db.transaction(async (tx) => {
    const [newBusiness] = await tx.insert(business).values(input).returning();

    await tx.insert(businessMembers).values({
      businessId: newBusiness.id,
      userId: session.id,
      role: "owner",
      canManageAds: true,
      canManageBilling: true,
      canManageBookings: true,
      canManageContent: true,
      canManageInventory: true,
      canManageLeads: true,
      canManageMembers: true,
      canManageSettings: true,
      canReadComments: true,
      canReadDms: true,
      canReplyComments: true,
      canReplyDms: true,
      canViewAnalytics: true,
      canViewBookings: true,
      canViewInventory: true,
      canViewLeads: true,
      canViewOrders: true,
    });

    await tx
      .update(user)
      .set({
        isOnboarded: true,
        role: "vendor",
      })
      .where(eq(user.id, session.id));

    return newBusiness;
  });
};

export const getByUserId = async (userId: string) => {
  const record = await db.query.businessMembers.findFirst({
    where: eq(businessMembers.userId, userId),
    with: {
      business: true,
    },
  });
  return record;
};
