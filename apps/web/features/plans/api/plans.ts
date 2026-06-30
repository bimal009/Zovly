import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import { Plan } from "@repo/types";

export const getPlans = async () => {
  const res = await axiosInstance.get<ApiResponse<Plan[]>>("/api/v1/plans");
  return res.data;
};
