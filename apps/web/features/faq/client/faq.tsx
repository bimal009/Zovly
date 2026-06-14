"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createFaq,
  deleteFaq,
  getFaqs,
  updateFaq,
  type CreateFaqInput,
  type ListFaqsParams,
  type UpdateFaqInput,
} from "../api/faq";

export const FAQS_KEY = ["faqs"] as const;

export const useGetFaqs = (params?: ListFaqsParams) => {
  return useQuery({
    queryKey: [...FAQS_KEY, params],
    queryFn: () => getFaqs(params),
  });
};

export const useCreateFaq = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateFaqInput) => createFaq(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};

export const useUpdateFaq = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateFaqInput }) =>
      updateFaq(id, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};

export const useDeleteFaq = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFaq(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};
