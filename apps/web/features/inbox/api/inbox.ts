import { axiosInstance } from "@/lib/axios";
import { PaginatedResponse, ApiResponse } from "@/lib/types/apiResponse";
import { Conversation, Message } from "../types/inbox";

export const getConversations = async (limit = 50, offset = 0) => {
  const res = await axiosInstance.get<PaginatedResponse<Conversation[]>>(
    `/api/v1/inbox/conversations?limit=${limit}&offset=${offset}`,
  );
  return res.data;
};

export const getMessages = async (conversationId: string, limit = 100) => {
  const res = await axiosInstance.get<ApiResponse<Message[]>>(
    `/api/v1/inbox/conversations/${conversationId}/messages?limit=${limit}`,
  );
  return res.data;
};
