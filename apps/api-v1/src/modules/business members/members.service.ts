import { and, eq } from "drizzle-orm";
import { db } from "../../config/db/db";
import { businessMembers } from "../../config/db/schema/members";

export const getByUserId = async (userId: string) => {
  const records = await db.query.businessMembers.findFirst({
    where: eq(businessMembers.userId, userId),
    with: { business: true },
  });
  return records;
};

export const getByBusinessAndUserId = async (
  businessId: string,
  userId: string,
) => {
  const record = await db.query.businessMembers.findFirst({
    where: and(
      eq(businessMembers.businessId, businessId),
      eq(businessMembers.userId, userId),
    ),
  });
  return record;
};
