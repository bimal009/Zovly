"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "@repo/ui/components/ui/sonner";
import {
  activateInstagram,
  getFacebookConnectionStatus,
  getInstagramConnectionStatus,
  subscribeInstagramWebhook,
  subscribeMessengerPage,
  toggleFacebookPage,
} from "../api/connections";

export const CONNECTIONS_KEY = ["connections"] as const;

function errMessage(error: unknown, fallback: string): string {
  const data = (error as { response?: { data?: { message?: string } } })?.response
    ?.data;
  return data?.message ?? fallback;
}



export function useFacebookConnectionStatus(businessId: string) {
  return useQuery({
    queryKey: ["facebook-connection-status", businessId],
    queryFn: () => getFacebookConnectionStatus(businessId),
    enabled: !!businessId,
  });
}

export function useToggleFacebookPage(businessId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ pageId, isActive }: { pageId: string; isActive: boolean }) =>
      toggleFacebookPage(businessId, pageId, isActive),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["facebook-connection-status", businessId] });
    },
  });
}

export const useInstagramConnectionStatus = () => {
  return useQuery({
    queryKey: [...CONNECTIONS_KEY, "instagram"],
    queryFn: getInstagramConnectionStatus,
  });
};



export const useSubscribeMessengerPage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (pageId: string) => subscribeMessengerPage(pageId),
    onSuccess: (res) => {
      toast.success(res.message);
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "facebook"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to enable Messenger")),
  });
};

export const useActivateInstagram = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: activateInstagram,
    onSuccess: (res) => {
      toast.success(res.message);
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "instagram"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to activate Instagram account")),
  });
};

export const useSubscribeInstagramWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: subscribeInstagramWebhook,
    onSuccess: (res) => {
      toast.success(res.message);
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "instagram"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to enable Instagram messaging")),
  });
};
