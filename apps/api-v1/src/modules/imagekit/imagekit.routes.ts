import { Router } from "express";
import { requireAuth } from "../../middlewares/auth.middleware";
import { getUploadSignature } from "./imagekit.controller";

const imagekitRouter = Router();

imagekitRouter.get("/auth", requireAuth, getUploadSignature);

export default imagekitRouter;
