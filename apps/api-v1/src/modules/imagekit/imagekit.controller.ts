import { Request, Response } from "express"
import { getImageKitSignature } from "./imagekit.service"
import { AppResponse } from "../../lib/response"

export const getUploadSignature = (req: Request, res: Response) => {
  const signature = getImageKitSignature()
  AppResponse.ok(res, signature, "Upload signature generated")
}
