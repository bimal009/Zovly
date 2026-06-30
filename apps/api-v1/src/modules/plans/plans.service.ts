import { db } from "../../config/db/db";

export const getPlans = async () => {
  const plans = await db.query.plans.findMany();
  return plans;
};
