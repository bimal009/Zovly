import { Request, Response, NextFunction } from "express";
import { ZodSchema } from "zod";

type Target = "body" | "query" | "params";

export const validate =
  (schema: ZodSchema, target: Target = "body") =>
  (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req[target]);
    if (!result.success) {
      res.status(400).json({
        success: false,
        errors: result.error.flatten().fieldErrors,
      });
      return;
    }
    Object.defineProperty(req, target, {
      value: result.data,
      writable: true,
      configurable: true,
    });
    next();
  };
