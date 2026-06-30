"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createBusiness } from "../api/business";
import type { CreateBusinessInput } from "@repo/types";

export const useCreateBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateBusinessInput) => createBusiness(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["business"] });
    },
  });
};
