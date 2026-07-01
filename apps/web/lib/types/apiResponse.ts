import { Meta } from "@repo/types";

export type ApiResponse<T> = {
  success: boolean;
  message: string;
  error?: string;
  data: T;
};



export type PaginatedResponse<T> = {
  success: boolean;
  message: string;
  error?: string;
  data: T;
  meta: Meta;
};
