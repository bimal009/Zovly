import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";

export type CreateBusinessInput = {
  name: string;
  description?: string;
  logo?: string;
  website?: string;
  phone: string;
  address?: string;
  city: string;
  country: string;
  type: "product" | "service" | "both";
};

export type BusinessResponse = {
  id: string;
  name: string;
  description: string | null;
  logo: string | null;
  website: string | null;
  phone: string | null;
  address: string | null;
  city: string | null;
  country: string;
  type: "product" | "service" | "both";
  created_at: string;
  updated_at: string;
};

export const createBusiness = async (input: CreateBusinessInput) => {
  const res = await axiosInstance.post<ApiResponse<BusinessResponse>>(
    "/api/v1/business/",
    input,
  );
  return res.data;
};
