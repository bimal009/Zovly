import { Response } from "express";

interface Meta {
  page?: number;
  limit?: number;
  total?: number;
  totalPages?: number;
}

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  errors?: unknown;
  meta?: Meta;
}

export class AppResponse {
  static ok<T>(res: Response, data: T, message = "Success") {
    return res.status(200).json({
      success: true,
      message,
      data,
    } satisfies ApiResponse<T>);
  }

  static created<T>(res: Response, data: T, message = "Created successfully") {
    return res.status(201).json({
      success: true,
      message,
      data,
    } satisfies ApiResponse<T>);
  }

  static paginated<T>(res: Response, data: T, meta: Meta, message = "Success") {
    return res.status(200).json({
      success: true,
      message,
      data,
      meta,
    } satisfies ApiResponse<T>);
  }

  static noContent(res: Response) {
    return res.status(204).send();
  }

  static badRequest(res: Response, errors: unknown, message = "Bad request") {
    return res.status(400).json({
      success: false,
      message,
      errors,
    } satisfies ApiResponse<never>);
  }

  static unauthorized(res: Response, message = "Unauthorized") {
    return res.status(401).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }

  static forbidden(res: Response, message = "Forbidden") {
    return res.status(403).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }

  static notFound(res: Response, message = "Not found") {
    return res.status(404).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }

  static conflict(res: Response, message = "Conflict") {
    return res.status(409).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }

  static unprocessable(
    res: Response,
    errors: unknown,
    message = "Validation failed",
  ) {
    return res.status(422).json({
      success: false,
      message,
      errors,
    } satisfies ApiResponse<never>);
  }

  static tooMany(res: Response, message = "Too many requests") {
    return res.status(429).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }

  static internal(res: Response, message = "Internal server error") {
    return res.status(500).json({
      success: false,
      message,
    } satisfies ApiResponse<never>);
  }
}
