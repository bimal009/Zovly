// business.service.ts
import { eq } from "drizzle-orm";
import { CreateBusinessInput } from "@repo/types";

import { db } from "../../config/db/db";
import { business } from "../../config/db/schema/business";
import { auth, User } from "../../lib/auth";
import { businessMembers } from "../../config/db/schema/members";
import { ConflictError, UnauthorizedError } from "../../lib/errors";

export const create = async (
  session: User,
  input: CreateBusinessInput,
  headers: Headers,
) => {
  if (!session) {
    throw new UnauthorizedError("User not found");
  }

  if (session.isOnboarded || session.role === "vendor") {
    throw new ConflictError("User already has a business");
  }

  const newBusiness = await db.transaction(async (tx) => {
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

    return newBusiness;
  });

  await auth.api.updateUser({
    body: {
      isOnboarded: true,
      role: "vendor",
    },
    headers,
  });

  return newBusiness;
};

export const getByUserId = async (userId: string) => {
  const record = await db.query.businessMembers.findFirst({
    where: eq(businessMembers.userId, userId),
    with: {
      business: true,
    },
  });

  return record ?? null;
};
