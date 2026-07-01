import { eq } from "drizzle-orm";
import { CreateBusinessInput } from "@repo/types";

import { db } from "../../config/db/db";
import { business } from "../../config/db/schema/business";
import { auth, User } from "../../lib/auth";
import { businessMembers } from "../../config/db/schema/members";
import {
  ConflictError,
  InternalServerError,
  NotFoundError,
  UnauthorizedError,
} from "../../lib/errors";
import { getByUserId } from "../business members/members.service";

export const create = async (
  session: User,
  input: CreateBusinessInput,
  headers: Headers,
) => {
  if (!session) {
    throw new UnauthorizedError("User not found");
  }

  const existingMember = await getByUserId(session.id);

  if (existingMember) {
    throw new ConflictError("User already has a business.");
  }
  console.log(existingMember);

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
      canEditBusiness:true,
      
    });

    try {
      await auth.api.updateUser({
        body: {
          isOnboarded: true,
          role: "vendor",
        },
        headers,
      });
    } catch (err) {
      throw new InternalServerError(
        "Failed to update user session. Please try again. Error: " + err,
      );
    }

    return newBusiness;
  });

  return newBusiness;
};

export const getById = async (id: string) => {
  const record = await db.query.business.findFirst({
    where: eq(business.id, id),
  });

  if (!record) {
    throw new NotFoundError("Business not found");
  }

  return record;
};
