"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "@repo/ui/components/ui/sonner";
import {
  activateInstagram,
  getBusinessAppConnections,
  getFacebookConnectionStatus,
  getInstagramConnectionStatus,
  subscribeInstagramWebhook,
  subscribeMessengerPage,
  toggleFacebookPage,
} from "../api/connections";

export const CONNECTIONS_KEY = ["connections"] as const;

export const useBusinessAppConnections = () => {
  return useQuery({
    queryKey: [...CONNECTIONS_KEY, "apps"],
    queryFn: getBusinessAppConnections,
  });
};

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
      toast.success(res.message);
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
