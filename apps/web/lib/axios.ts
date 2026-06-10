import axios from "axios";
import { authClient } from "./auth-client";

export const axiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

async function getAuthToken() {
  const { data } = await authClient.getSession();
  return data?.session.token ?? null;
}

axiosInstance.interceptors.request.use(async (config) => {
  const token = await getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
