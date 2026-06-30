import axios from "axios";

export const axiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_BETTER_AUTH_URL,
  withCredentials: true,
});
