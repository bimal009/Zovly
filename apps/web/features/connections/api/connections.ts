import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import { FacebookConnectionStatus } from "../types/connections";

export const getFacebookConnectionStatus = async () => {
  const res = await axiosInstance.get<ApiResponse<FacebookConnectionStatus>>(
    "/api/v1/connections/facebook",
  );
  return res.data;
};

export const getFacebookConnectURL = async (): Promise<string> => {
  const res = await axiosInstance.get<ApiResponse<{ url: string }>>(
    "/api/v1/auth/facebook/connect",
  );
  return res.data.data.url;
};

export const toggleFacebookPage = async (pageId: string) => {
  const res = await axiosInstance.patch<ApiResponse<{ is_active: boolean }>>(
    `/api/v1/connections/facebook/pages/${pageId}/toggle`,
  );
  return res.data;
};
