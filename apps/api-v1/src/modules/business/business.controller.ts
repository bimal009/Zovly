import { Request, Response } from "express";
import { create, getById } from "./business.service";
import { User } from "../../lib/auth";
import { AppResponse } from "../../lib/response";
import { handleError } from "../../lib/errors";
import { getByUserId } from "../business members/members.service";

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
    const business = await getById(req.user?.id as string);

    return AppResponse.ok(res, business, "User business fetched successfully");
  } catch (error) {
    return handleError(res, "getBusinessByUserId", error);
  }
};

export const getBusinessById = async (req: Request, res: Response) => {
  try {
    const business = await getById(req.user?.id as string);
    return AppResponse.ok(res, business, "Business fetched successfully");
  } catch (error) {
    return handleError(res, "getBusinessById", error);
  }
};
