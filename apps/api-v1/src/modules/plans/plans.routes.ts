import { Router } from "express";
import { getPlansController } from "./plans.controller";

export const plansRoutes = Router();

plansRoutes.get("/", getPlansController);
