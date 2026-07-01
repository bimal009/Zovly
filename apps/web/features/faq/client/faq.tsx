"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createFaq,
  deleteFaq,
  getFaqs,
  updateFaq,
} from "../api/faq";
import { CreateFaqInput, PaginationQuery, UpdateFaqInput } from "@repo/types";

export const FAQS_KEY = ["faqs"] as const;

export const useGetFaqs = (
  businessId?: string,
  params?: PaginationQuery,
) => {
  return useQuery({
    queryKey: [...FAQS_KEY, businessId, params],
    queryFn: () => getFaqs(businessId!, params),
    enabled: !!businessId,
  });
};

export const useCreateFaq = (businessId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateFaqInput) => createFaq(businessId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};

export const useUpdateFaq = (businessId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateFaqInput }) =>
      updateFaq(businessId, id, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};

export const useDeleteFaq = (businessId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFaq(businessId, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: FAQS_KEY }),
  });
};
