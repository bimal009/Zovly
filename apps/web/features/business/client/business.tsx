"use client";

import { useQuery } from "@tanstack/react-query";
import { getBusiness } from "../api/business";

export const useGetBusiness = (enabled = true) => {
  return useQuery({
    queryKey: ["business"],
    queryFn: getBusiness,
    enabled,
    retry: false,
  });
};
