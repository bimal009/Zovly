import { axiosInstance } from "@/lib/axios";
import { ApiResponse, PaginatedResponse } from "@/lib/types/apiResponse";
import {
  CreateProductInput,
  ListProductsParams,
  Product,
  UpdateProductInput,
} from "../types/products";

export const getProducts = async (params?: ListProductsParams) => {
  const res = await axiosInstance.get<PaginatedResponse<Product[]>>(
    "/api/v1/products",
    { params },
  );
  return res.data;
};

export const getProductById = async (id: string) => {
  const res = await axiosInstance.get<ApiResponse<Product>>(
    `/api/v1/products/${id}`,
  );
  return res.data;
};

export const getLowStockProducts = async () => {
  const res = await axiosInstance.get<ApiResponse<Product[]>>(
    "/api/v1/products/low-stock",
  );
  return res.data;
};

export const createProduct = async (input: CreateProductInput) => {
  const res = await axiosInstance.post<ApiResponse<Product>>(
    "/api/v1/products",
    input,
  );
  return res.data;
};

export const updateProduct = async (id: string, input: UpdateProductInput) => {
  const res = await axiosInstance.patch<ApiResponse<Product>>(
    `/api/v1/products/${id}`,
    input,
  );
  return res.data;
};

export const deleteProduct = async (id: string) => {
  const res = await axiosInstance.delete<ApiResponse<null>>(
    `/api/v1/products/${id}`,
  );
  return res.data;
};
