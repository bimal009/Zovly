import { Request, Response } from "express";
import { facebookWebhookChallenge, processFacebookWebhookPayload } from "./facebook-webhooks.service";

export async function verifyFacebookWebhookChallenge(
  req: Request,
  res: Response
): Promise<void> {
  const hubMode = req.query["hub.mode"];
  const hubVerifyToken = req.query["hub.verify_token"];
  const hubChallenge = req.query["hub.challenge"];

  try {
    const challenge = facebookWebhookChallenge(hubMode, hubVerifyToken, hubChallenge);
    res.status(200).send(challenge);
  } catch {
    res.sendStatus(403);
  }
}

export const handleFacebookWebhook = async (req: Request, res: Response): Promise<void> => {
  console.log(req.body)
  const webhookObject = req.body?.object;
  res.sendStatus(200);

  if (webhookObject !== "page") {
    console.warn(`Unhandled Facebook webhook object type: ${webhookObject}`);
    return;
  }

  try {
    await processFacebookWebhookPayload(req.body);
  } catch (err) {
    console.error("Failed to process Facebook webhook payload:", err);
  }
};