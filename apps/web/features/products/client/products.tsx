"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createProduct,
  deleteProduct,
  getLowStockProducts,
  getProducts,
  updateProduct,
} from "../api/products";
import {
  type CreateProductInput,
  type ListProductsParams,
  type UpdateProductInput,
} from "../types/products";

export const PRODUCTS_KEY = ["products"] as const;

export const useGetProducts = (params?: ListProductsParams) => {
  return useQuery({
    queryKey: [...PRODUCTS_KEY, params],
    queryFn: () => getProducts(params),
  });
};

export const useGetLowStockProducts = () => {
  return useQuery({
    queryKey: [...PRODUCTS_KEY, "low-stock"],
    queryFn: getLowStockProducts,
  });
};

export const useCreateProduct = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateProductInput) => createProduct(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: PRODUCTS_KEY }),
  });
};

export const useUpdateProduct = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateProductInput }) =>
      updateProduct(id, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: PRODUCTS_KEY }),
  });
};

export const useDeleteProduct = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProduct(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: PRODUCTS_KEY }),
  });
};
