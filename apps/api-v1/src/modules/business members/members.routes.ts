import { Router } from "express";
import { requireAuth } from "../../middlewares/auth.middleware";
import { validate } from "../../middlewares/validate.middleware";
import { createBusinessSchema } from "@repo/types";
import { getMembersByUserId } from "./members.controller";

const businessRouter = Router();

businessRouter.get("/", requireAuth, getMembersByUserId);

export default businessRouter;
