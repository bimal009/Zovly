import { getPlans } from "./plans.service";
import { AppResponse } from "../../lib/response";
import { Request, Response } from "express";

export const getPlansController = async (req: Request, res: Response) => {
  const plans = await getPlans();
  return AppResponse.ok(res, plans, "Plans fetched successfully");
};
