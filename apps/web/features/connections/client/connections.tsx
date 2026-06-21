"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "@repo/ui/components/ui/sonner";
import {
  getFacebookConnectionStatus,
  getInstagramConnectionStatus,
  subscribeInstagramWebhook,
  subscribeMessengerPage,
  toggleFacebookPage,
} from "../api/connections";

export const CONNECTIONS_KEY = ["connections"] as const;

// Pulls a human-readable message out of an axios/API error, falling back to a default.
function errMessage(error: unknown, fallback: string): string {
  const data = (error as { response?: { data?: { message?: string } } })?.response
    ?.data;
  return data?.message ?? fallback;
}

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
    onSuccess: (res) => {
      toast.success(
        res.data?.is_active ? "Page enabled" : "Page disabled",
      );
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "facebook"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to update page")),
  });
};

export const useInstagramConnectionStatus = () => {
  return useQuery({
    queryKey: [...CONNECTIONS_KEY, "instagram"],
    queryFn: getInstagramConnectionStatus,
  });
};

export const useMessengerConnectionStatus = () => {
  return useQuery({
    queryKey: [...CONNECTIONS_KEY, "facebook"],
    queryFn: getFacebookConnectionStatus,
  });
};

export const useSubscribeMessengerPage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (pageId: string) => subscribeMessengerPage(pageId),
    onSuccess: () => {
      toast.success("Messenger enabled — you'll now receive messages");
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "facebook"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to enable Messenger")),
  });
};

export const useSubscribeInstagramWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: subscribeInstagramWebhook,
    onSuccess: () => {
      toast.success("Instagram messaging enabled — you'll now receive DMs");
      qc.invalidateQueries({ queryKey: [...CONNECTIONS_KEY, "instagram"] });
    },
    onError: (error) =>
      toast.error(errMessage(error, "Failed to enable Instagram messaging")),
  });
};
