"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "@repo/ui/components/ui/sonner";
import { createCategory, getCategories } from "../api/categories";
import { type CreateCategoryInput } from "../types/categories";

export const CATEGORIES_KEY = ["categories"] as const;

// Pulls a human-readable message out of an axios/API error, falling back to a default.
function errMessage(error: unknown, fallback: string): string {
  const data = (error as { response?: { data?: { message?: string } } })
    ?.response?.data;
  return data?.message ?? fallback;
}

export const useGetCategories = () => {
  return useQuery({
    queryKey: CATEGORIES_KEY,
    queryFn: getCategories,
  });
};

export const useCreateCategory = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateCategoryInput) => createCategory(input),
    onSuccess: (res) => {
      toast.success(res.message);
      qc.invalidateQueries({ queryKey: CATEGORIES_KEY });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to create category")),
  });
};
