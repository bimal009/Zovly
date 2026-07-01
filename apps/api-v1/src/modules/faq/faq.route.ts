import { Router } from "express";
import { requireAuth } from "../../middlewares/auth.middleware";
import {
  businessEditAuthorization,
  validateBusiness,
  validateBusinessMember,
} from "../../middlewares/business.middleware";
import { validate } from "../../middlewares/validate.middleware";
import {
  createFaqSchema,
  updateFaqSchema,
  PaginationQuerySchema,
} from "@repo/types";
import { createFAQ, getFAQs, updateFAQ, deleteFAQ } from "./faq.controller";

const faqRouter = Router();

faqRouter.post(
  "/:id",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessEditAuthorization,
  validate(createFaqSchema, "body"),
  createFAQ,
);

faqRouter.get(
  "/:id",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  validate(PaginationQuerySchema, "query"),
  getFAQs,
);

faqRouter.patch(
  "/:id/:faqId",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessEditAuthorization,
  validate(updateFaqSchema, "body"),
  updateFAQ,
);

faqRouter.delete(
  "/:id/:faqId",
  requireAuth,
  validateBusiness,
  validateBusinessMember,
  businessEditAuthorization,
  deleteFAQ,
);

export default faqRouter;