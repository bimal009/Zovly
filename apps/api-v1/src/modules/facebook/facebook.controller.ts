import { Request, Response } from "express";

import { AppResponse } from "../../lib/response";
import { handleError, ValidationError } from "../../lib/errors";
import { buildFacebookConnectUrl, handleFacebookCallback, listFacebookPages, toggleFacebookPage } from "./facebook.service";

export async function connectFacebook(req: Request, res: Response) {
  try {
    const { id: businessId } = req.params;
    const url = await buildFacebookConnectUrl(businessId as string);
    return AppResponse.ok(res, { url }, "facebook oauth url");
  } catch (err) {
    return handleError(res, "connectFacebook", err);
  }
}

export async function connectFacebookCallback(req: Request, res: Response) {
  const { code, state } = req.query;
  try {
    const pages = await handleFacebookCallback(state as string, code as string);
    return res.redirect(
      `${process.env.FRONTEND_URL}/settings/integrations?fb=connected&count=${pages.length}`
    );
  } catch (err) {
    console.error("connectFacebookCallback error:", err);
    return res.redirect(`${process.env.FRONTEND_URL}/settings/integrations?fb=error`);
  }
}

export async function listFacebookPagesHandler(req: Request, res: Response) {
  try {
    const { id: businessId } = req.params;
    const result = await listFacebookPages(businessId as string);
    return AppResponse.ok(res, result, "facebook connection status");
  } catch (err) {
    return handleError(res, "listFacebookPagesHandler", err);
  }
}

export async function toggleFacebookPageHandler(req: Request, res: Response) {
  try {
    const { id: businessId, pageId } = req.params;
    const { isActive } = req.body;

    if (typeof isActive !== "boolean") {
      throw new ValidationError("isActive (boolean) is required in the request body.");
    }

    const updated = await toggleFacebookPage(businessId as string, pageId as string, isActive);
    return AppResponse.ok(res, { page: updated }, isActive ? "Page connected" : "Page disconnected");
  } catch (err) {
    return handleError(res, "toggleFacebookPageHandler", err);
  }
}