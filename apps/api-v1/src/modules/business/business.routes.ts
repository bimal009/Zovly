import { Router } from "express";
import { requireAuth } from "../../middlewares/auth.middleware";
import { validate } from "../../middlewares/validate.middleware";
import { createBusinessSchema } from "@repo/types";
import { createBusiness } from "./business.controller";

const businessRouter = Router();

businessRouter.post(
  "/",
  requireAuth,
  validate(createBusinessSchema, "body"),
  createBusiness,
);
