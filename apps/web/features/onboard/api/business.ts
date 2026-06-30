import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import type { CreateBusinessInput } from "@repo/types";

export type BusinessResponse = {
  id: string;
  name: string;
  description: string;
  logo: string | null;
  website: string | null;
  phone: string | null;
  address: string;
  city: string | null;
  country: string;
  type: "product" | "service" | "both";
  created_at: string;
  updated_at: string;
};

export const createBusiness = async (input: CreateBusinessInput) => {
  const res = await axiosInstance.post<ApiResponse<BusinessResponse>>(
    "/api/v1/business",
    input,
  );
  return res.data;
};
