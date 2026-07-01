import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import {
  FacebookConnectionStatus,
  InstagramConnectionStatus,
} from "../types/connections";




export const getInstagramConnectionStatus = async () => {
  const res = await axiosInstance.get<ApiResponse<InstagramConnectionStatus>>(
    "/api/v1/connections/instagram",
  );
  return res.data;
};

export const getInstagramConnectURL = async (): Promise<string> => {
  const res = await axiosInstance.get<ApiResponse<{ url: string }>>(
    "/api/v1/auth/instagram/connect",
  );
  return res.data.data.url;
};

export const subscribeMessengerPage = async (pageId: string) => {
  const res = await axiosInstance.post<ApiResponse<null>>(
    `/api/v1/connections/messenger/pages/${pageId}/subscribe`,
  );
  return res.data;
};

export const activateInstagram = async () => {
  const res = await axiosInstance.post<ApiResponse<{ is_active: boolean }>>(
    "/api/v1/connections/instagram/activate",
  );
  return res.data;
};

export const subscribeInstagramWebhook = async () => {
  const res = await axiosInstance.post<ApiResponse<null>>(
    "/api/v1/connections/instagram/subscribe",
  );
  return res.data;
};






export const getFacebookConnectURL = async (businessId: string): Promise<string> => {
  const res = await axiosInstance.get<ApiResponse<{ url: string }>>(
    `/api/v1/apps/connect/facebook/${businessId}`
  );
  return res.data.data.url;
};

export const getFacebookConnectionStatus = async (
  businessId: string
): Promise<ApiResponse<FacebookConnectionStatus>> => {
  const res = await axiosInstance.get<ApiResponse<FacebookConnectionStatus>>(
    `/api/v1/apps/connect/facebook/${businessId}/pages`
  );
  return res.data;
};

export const toggleFacebookPage = async (
  businessId: string,
  pageId: string,
  isActive: boolean
) => {
  const res = await axiosInstance.patch(
    `/api/v1/apps/connect/facebook/${businessId}/pages/${pageId}/toggle`,
    { isActive }
  );
  return res.data;
};