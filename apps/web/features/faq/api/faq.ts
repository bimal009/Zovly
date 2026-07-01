import { axiosInstance } from "@/lib/axios";
import { ApiResponse, PaginatedResponse } from "@/lib/types/apiResponse";
import { CreateFaqInput, Faq, PaginationQuery, UpdateFaqInput } from "@repo/types";

export const getFaqs = async (businessId: string, params?: PaginationQuery) => {
  const res = await axiosInstance.get<PaginatedResponse<Faq[]>>(
    `/api/v1/faq/${businessId}`,
    { params },
  );
  return res.data;
};

export const createFaq = async (businessId: string, input: CreateFaqInput) => {
  const res = await axiosInstance.post<ApiResponse<Faq>>(
    `/api/v1/faq/${businessId}`,
    input,
  );
  return res.data;
};

export const updateFaq = async (
  businessId: string,
  id: string,
  input: UpdateFaqInput,
) => {
  const res = await axiosInstance.patch<ApiResponse<Faq>>(
    `/api/v1/faq/${businessId}/${id}`,
    input,
  );
  return res.data;
};

export const deleteFaq = async (businessId: string, id: string) => {
  const res = await axiosInstance.delete<ApiResponse<null>>(
    `/api/v1/faq/${businessId}/${id}`,
  );
  return res.data;
};