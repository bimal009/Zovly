"use client";

import { useQuery } from "@tanstack/react-query";
import { getPlans } from "../api/plans";

export const useGetPlans = () => {
  return useQuery({
    queryKey: ["plans"],
    queryFn: getPlans,
  });
};
