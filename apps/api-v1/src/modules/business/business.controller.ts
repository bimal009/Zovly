import { Request, Response } from "express";
import { create, getByUserId } from "./business.service";
import { User } from "../../lib/auth";
import { AppResponse } from "../../lib/response";
import { handleError } from "../../lib/errors";

export const createBusiness = async (req: Request, res: Response) => {
  try {
    const business = await create(
      req.user as User,
      req.body,
      req.headers as unknown as Headers,
    );

    return AppResponse.created(res, business, "Business created successfully");
  } catch (error) {
    return handleError(res, "createBusiness", error);
  }
};

export const getBusinessByUserId = async (req: Request, res: Response) => {
  try {
    const business = await getByUserId(req.user?.id as string);

    return AppResponse.ok(res, business, "User business fetched successfully");
  } catch (error) {
    return handleError(res, "getBusinessByUserId", error);
  }
};
