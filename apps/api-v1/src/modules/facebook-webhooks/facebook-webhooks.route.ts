import { Router } from "express";
import { handleFacebookWebhook, verifyFacebookWebhookChallenge } from "./facebook-webhooks.controller";

const facebookWebhookRouter = Router();


facebookWebhookRouter.get("/",verifyFacebookWebhookChallenge);
facebookWebhookRouter.post("/",handleFacebookWebhook);

export default facebookWebhookRouter;