import { NextFunction, Request, Response } from "express";
import { getById } from "../modules/business/business.service";
import {
  BadRequestError,
  ForbiddenError,
  UnauthorizedError,
} from "../lib/errors";
import { User } from "../lib/auth";
import { getByBusinessAndUserId } from "../modules/business members/members.service";

export const validateBusiness = async (
  req: Request,
  res: Response,
  next: NextFunction,
) => {
  const user = req.user as User;
  if (!user) {
    return next(new UnauthorizedError("User not found"));
  }

  if (!req.params.id) {
    return next(new BadRequestError("Business ID is required"));
  }

  const business = await getById(req.params.id as string);
  req.business = business;

  next();
};

export const validateBusinessMember = async (
  req: Request,
  res: Response,
  next: NextFunction,
) => {
  const user = req.user as User;
  if (!user) {
    return next(new UnauthorizedError("User not found"));
  }

  if (!req.params.id) {
    return next(new BadRequestError("Business ID is required"));
  }

  const member = await getByBusinessAndUserId(req.params.id as string, user.id);
  if (!member) {
    return next(new UnauthorizedError("You are not a member of this business"));
  }

  req.member = member;

  next();
};

export const businessEditAuthorization = async (
  req: Request,
  res: Response,
  next: NextFunction,
) => {
  const user = req.user as User;
  if (!user) {
    return next(new UnauthorizedError("User not found"));
  }

  if (!req.params.id) {
    return next(new BadRequestError("Business ID is required"));
  }

  const member = await getByBusinessAndUserId(req.params.id as string, user.id);
  if (!member) {
    return next(new UnauthorizedError("You are not a member of this business"));
  }

  if (!member.canEditBusiness) {
    return next(new ForbiddenError("You can't edit this business's details"));
  }
  req.member = member;

  next();
};
