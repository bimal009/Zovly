import { Request, Response } from "express";
import { AppResponse } from "../../lib/response";
import { handleError } from "../../lib/errors";
import { create, get, update, remove } from "./faq.service";
import { Meta, MetaSchema } from "@repo/types";

export const createFAQ = async (req: Request, res: Response) => {
  try {
    const faq = await create(req.body, req.params.id as string);
    return AppResponse.created(res, faq, "FAQ created successfully");
  } catch (error) {
    return handleError(res, "Faq creation", error);
  }
};

export const getFAQs = async (req: Request, res: Response) => {
  try {
    const query = {
      page: Number(req.query.page) || 1,
      limit: Number(req.query.limit) || 10,
      search: req.query.search as string | undefined,
    };

    const faqs = await get(query, req.params.id as string);
    const meta: Meta = {
      limit: faqs.limit,
      page: faqs.page,
      total: faqs.total,
      totalPages: faqs.totalPages,
    };
    return AppResponse.paginated(res, faqs.data, meta);
  } catch (error) {
    return handleError(res, "Faq fetch", error);
  }
};

export const updateFAQ = async (req: Request, res: Response) => {
  try {
    const faq = await update(
      req.params.faqId as string,
      req.body,
      req.params.id as string,
    );
    return AppResponse.ok(res, faq, "FAQ updated successfully");
  } catch (error) {
    return handleError(res, "Faq update", error);
  }
};

export const deleteFAQ = async (req: Request, res: Response) => {
  try {
    const faq = await remove(req.params.faqId as string, req.params.id as string);
    return AppResponse.ok(res, faq, "FAQ deleted successfully");
  } catch (error) {
    return handleError(res, "Faq deletion", error);
  }
};