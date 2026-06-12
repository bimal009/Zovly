import { axiosInstance } from "@/lib/axios";

export async function fetchIKAuth() {
  const { data } = await axiosInstance.get<{
    data: { signature: string; expire: number; token: string };
  }>("/api/v1/images/auth");

  return data.data;
}
