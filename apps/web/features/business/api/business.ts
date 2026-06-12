import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";

export type BusinessResponse = {
  id: string;
  name: string;
  description: string | null;
  logo: string | null;
  website: string | null;
  phone: string | null;
  address: string | null;
  city: string | null;
  country: string;
  type: "product" | "service" | "both";
  created_at: string;
  updated_at: string;
};

export type MemberResponse = {
  id: string;
  business_id: string;
  user_id: string;
  role: string;
  joined_at: string;
  last_seen_at: string | null;
  created_at: string;
  updated_at: string;
  can_manage_ads: boolean;
  can_manage_billing: boolean;
  can_manage_bookings: boolean;
  can_manage_content: boolean;
  can_manage_inventory: boolean;
  can_manage_leads: boolean;
  can_manage_members: boolean;
  can_manage_settings: boolean;
  can_read_comments: boolean;
  can_read_dms: boolean;
  can_reply_comments: boolean;
  can_reply_dms: boolean;
  can_view_analytics: boolean;
  can_view_bookings: boolean;
  can_view_inventory: boolean;
  can_view_leads: boolean;
  can_view_orders: boolean;
};

export type GetBusinessResponse = {
  Business: BusinessResponse;
  Member: MemberResponse;
};

export const getBusiness = async () => {
  const res = await axiosInstance.get<ApiResponse<GetBusinessResponse>>(
    "/api/v1/business/",
  );
  return res.data;
};
