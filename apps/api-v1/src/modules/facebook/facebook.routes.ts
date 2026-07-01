import { Router } from "express";
import {
  connectFacebook,
  connectFacebookCallback,
  listFacebookPagesHandler,
  toggleFacebookPageHandler,
} from "./facebook.controller";
import { requireAuth } from "../../middlewares/auth.middleware";
import { businessSettingsAuthorization, validateBusiness, validateBusinessMember } from "../../middlewares/business.middleware";

const facebookRouter = Router();

facebookRouter.get("/connect/facebook/callback", connectFacebookCallback);

facebookRouter.get(
  "/connect/facebook/:id",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessSettingsAuthorization,
  connectFacebook
);

facebookRouter.get(
  "/connect/facebook/:id/pages",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessSettingsAuthorization,
  listFacebookPagesHandler
);

facebookRouter.patch(
  "/connect/facebook/:id/pages/:pageId/toggle",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessSettingsAuthorization,
  toggleFacebookPageHandler
);

export default facebookRouter;