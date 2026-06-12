import { axiosInstance } from "@/lib/axios";
import { ApiResponse, PaginatedResponse } from "@/lib/types/apiResponse";

// ─── Domain types (match Go JSON tags exactly) ────────────────────────────────

export type ServiceType = "appointment" | "membership" | "class" | "package";
export type ServiceStatus = "active" | "inactive" | "archived";
export type BillingInterval = "weekly" | "monthly" | "quarterly" | "yearly";

export type ServiceFeature = {
  label: string;
  value: string;
};

export type Service = {
  id: string;
  business_id: string;
  type: ServiceType;
  status: ServiceStatus;
  name: string;
  description: string | null;
  price: number;
  cost_price: number | null;
  mrp: number | null;
  currency: string;
  requires_deposit: boolean;
  deposit_amount: number | null;
  duration_min: number | null;
  buffer_min: number | null;
  max_advance_days: number | null;
  google_calendar_id: string | null;
  max_concurrent: number | null;
  billing_interval: BillingInterval | null;
  trial_days: number | null;
  session_count: number | null;
  validity_days: number | null;
  features: ServiceFeature[];
  images: string[];
  created_at: string;
  updated_at: string;
};

export type CreateServiceInput = {
  type: ServiceType;
  name: string;
  status?: ServiceStatus;
  description?: string;
  price: number;
  cost_price?: number;
  mrp?: number;
  currency?: string;
  requires_deposit?: boolean;
  deposit_amount?: number;
  duration_min?: number;
  buffer_min?: number;
  max_advance_days?: number;
  google_calendar_id?: string;
  max_concurrent?: number;
  billing_interval?: BillingInterval;
  trial_days?: number;
  session_count?: number;
  validity_days?: number;
  features?: ServiceFeature[];
  images?: string[];
};

export type UpdateServiceInput = {
  status?: ServiceStatus;
  name?: string;
  description?: string | null;
  price?: number;
  cost_price?: number | null;
  mrp?: number | null;
  currency?: string;
  requires_deposit?: boolean;
  deposit_amount?: number | null;
  duration_min?: number | null;
  buffer_min?: number | null;
  max_advance_days?: number | null;
  google_calendar_id?: string | null;
  max_concurrent?: number | null;
  billing_interval?: BillingInterval | null;
  trial_days?: number | null;
  session_count?: number | null;
  validity_days?: number | null;
  features?: ServiceFeature[];
  images?: string[];
};

export type ListServicesParams = {
  type?: ServiceType;
  status?: ServiceStatus;
  search?: string;
  limit?: number;
  offset?: number;
};

// ─── API functions ─────────────────────────────────────────────────────────────

export const getServices = async (params?: ListServicesParams) => {
  const res = await axiosInstance.get<PaginatedResponse<Service[]>>(
    "/api/v1/services",
    { params }
  );
  return res.data;
};

export const getServiceById = async (id: string) => {
  const res = await axiosInstance.get<ApiResponse<Service>>(
    `/api/v1/services/${id}`
  );
  return res.data;
};

export const createService = async (input: CreateServiceInput) => {
  const res = await axiosInstance.post<ApiResponse<Service>>(
    "/api/v1/services",
    input
  );
  return res.data;
};

export const updateService = async (id: string, input: UpdateServiceInput) => {
  const res = await axiosInstance.patch<ApiResponse<Service>>(
    `/api/v1/services/${id}`,
    input
  );
  return res.data;
};

export const deleteService = async (id: string) => {
  const res = await axiosInstance.delete<ApiResponse<null>>(
    `/api/v1/services/${id}`
  );
  return res.data;
};
