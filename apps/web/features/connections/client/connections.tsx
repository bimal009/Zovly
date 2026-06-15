"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getFacebookConnectionStatus,
  toggleFacebookPage,
} from "../api/connections";

export const CONNECTIONS_KEY = ["connections"] as const;

export const useFacebookConnectionStatus = () => {
  return useQuery({
    queryKey: [...CONNECTIONS_KEY, "facebook"],
    queryFn: getFacebookConnectionStatus,
  });
};

export const useToggleFacebookPage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (pageId: string) => toggleFacebookPage(pageId),
    onSuccess: () => qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "facebook"] }),
  });
};
