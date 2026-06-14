import { axiosInstance } from "@/lib/axios";
import { ApiResponse, PaginatedResponse } from "@/lib/types/apiResponse";

export type Faq = {
  id: string;
  business_id: string;
  question: string;
  answer: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
};

export type CreateFaqInput = {
  question: string;
  answer: string;
};

export type UpdateFaqInput = {
  question?: string;
  answer?: string;
  is_active?: boolean;
  sort_order?: number;
};

export type ListFaqsParams = {
  search?: string;
  limit?: number;
  offset?: number;
};

export const getFaqs = async (params?: ListFaqsParams) => {
  const res = await axiosInstance.get<PaginatedResponse<Faq[]>>(
    "/api/v1/faqs/all",
    { params },
  );
  return res.data;
};

export const createFaq = async (input: CreateFaqInput) => {
  const res = await axiosInstance.post<ApiResponse<Faq>>(
    "/api/v1/faqs/create",
    input,
  );
  return res.data;
};

export const updateFaq = async (id: string, input: UpdateFaqInput) => {
  const res = await axiosInstance.patch<ApiResponse<Faq>>(
    `/api/v1/faqs/${id}`,
    input,
  );
  return res.data;
};

export const deleteFaq = async (id: string) => {
  const res = await axiosInstance.delete<ApiResponse<null>>(
    `/api/v1/faqs/${id}`,
  );
  return res.data;
};
