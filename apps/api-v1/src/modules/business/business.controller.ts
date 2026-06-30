import { Request, Response } from "express";
import { create } from "./business.service";
import { User } from "../../lib/auth";
import { AppResponse } from "../../lib/response";

export const createBusiness = async (req: Request, res: Response) => {
  const business = create(req.user as User, req.body);
  AppResponse.created(res, business, "Business created sucessfully");
};
