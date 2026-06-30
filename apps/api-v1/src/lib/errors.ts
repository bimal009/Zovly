import { Response } from "express";
import { AppResponse } from "./response";

export class AppError extends Error {
  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
  }
}

export class ValidationError extends AppError {}
export class ConflictError extends AppError {}
export class NotFoundError extends AppError {}
export class UnauthorizedError extends AppError {}
export class ForbiddenError extends AppError {}

export const handleError = (res: Response, label: string, error: unknown) => {
  console.error(`${label} error:`, error);

  if (error instanceof ConflictError) {
    return AppResponse.conflict(res, error.message);
  }
  if (error instanceof NotFoundError) {
    return AppResponse.notFound(res, error.message);
  }
  if (error instanceof UnauthorizedError) {
    return AppResponse.unauthorized(res, error.message);
  }
  if (error instanceof ValidationError) {
    return AppResponse.unprocessable(res, error.message);
  }

  return AppResponse.badRequest(
    res,
    error instanceof Error ? error.message : "Internal Server Error",
  );
};
