import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import { Category, CreateCategoryInput } from "../types/categories";

export const getCategories = async () => {
  const res = await axiosInstance.get<ApiResponse<Category[]>>(
    "/api/v1/categories",
  );
  return res.data;
};

export const getCategoryById = async (id: string) => {
  const res = await axiosInstance.get<ApiResponse<Category>>(
    `/api/v1/categories/${id}`,
  );
  return res.data;
};

export const createCategory = async (input: CreateCategoryInput) => {
  const res = await axiosInstance.post<ApiResponse<null>>(
    "/api/v1/categories",
    input,
  );
  return res.data;
};
