import { Request, Response, NextFunction } from "express"
import { fromNodeHeaders } from "better-auth/node"
import { auth } from "../lib/auth"

export async function requireAuth(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const session = await auth.api.getSession({
      headers: fromNodeHeaders(req.headers),
    })

    if (!session || !session.user) {
      return res.status(401).json({ error: "Unauthorized access" })
    }

    req.user = session.user
    req.session = session.session

    next()
  } catch (error) {
    return res.status(500).json({ error: "Internal authentication error" })
  }
}
