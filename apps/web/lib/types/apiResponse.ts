export type ApiResponse<T> = {
  success: boolean;
  message: string;
  error?: string;
  data: T;
};

export type PaginatedMeta = {
  total: number;
  limit: number;
  offset: number;
};

export type PaginatedResponse<T> = {
  success: boolean;
  message: string;
  error?: string;
  data: T;
  meta: PaginatedMeta;
};
