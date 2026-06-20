"use client";

import { useQuery } from "@tanstack/react-query";
import { getConversations, getMessages } from "../api/inbox";

export const INBOX_KEY = ["inbox"] as const;

export const useConversations = () => {
  return useQuery({
    queryKey: [...INBOX_KEY, "conversations"],
    queryFn: () => getConversations(),
    staleTime: 5000,
    refetchInterval: 5000,
  });
};

export const useMessages = (conversationId: string | null) => {
  return useQuery({
    queryKey: [...INBOX_KEY, "messages", conversationId],
    queryFn: () => getMessages(conversationId!),
    enabled: !!conversationId,
    staleTime: 5000,
    refetchInterval: 5000,
  });
};
