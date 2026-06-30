import { Router } from "express";
import { requireAuth } from "../../middlewares/auth.middleware";
import { validate } from "../../middlewares/validate.middleware";
import { createBusinessSchema } from "@repo/types";
import { createBusiness, getBusinessByUserId } from "./business.controller";

const businessRouter = Router();

businessRouter.post(
  "/",
  requireAuth,
  validate(createBusinessSchema, "body"),
  createBusiness,
);

businessRouter.get("/", requireAuth, getBusinessByUserId);

export default businessRouter;
