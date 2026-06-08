import { axiosInstance } from "@/lib/axios";
import { ApiResponse } from "@/lib/types/apiResponse";
import { Plan } from "@/lib/types/plans";

export const getPlans = async () => {
  const res = await axiosInstance.get<ApiResponse<Plan[]>>("/api/v1/plans/all");
  return res.data;
};
