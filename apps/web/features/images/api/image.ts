import { authClient } from "@/lib/auth-client";
import { axiosInstance } from "@/lib/axios";

export async function fetchIKAuth() {
  const session = await authClient.getSession();
  const token = session.data?.session.token;

  const { data } = await axiosInstance.get<{
    data: { signature: string; expire: number; token: string };
  }>("/api/v1/images/auth", {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  return data.data;
}
