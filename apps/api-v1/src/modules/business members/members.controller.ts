import { Request, Response } from "express";
import { AppResponse } from "../../lib/response";
import { handleError } from "../../lib/errors";
import { getByUserId } from "./members.service";

export const getMembersByUserId = async (req: Request, res: Response) => {
  try {
    const members = await getByUserId(req.user?.id as string);

    return AppResponse.ok(res, members, "User members fetched successfully");
  } catch (error) {
    return handleError(res, "getMembersByUserId", error);
  }
};
