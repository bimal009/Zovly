import { z } from "zod";

export const ProductStatusSchema = z.enum(["active", "inactive", "archived"]);
export type ProductStatus = z.infer<typeof ProductStatusSchema>;

export type Product = {
  id: string;
  business_id: string;
  name: string;
  description: string | null;
  sku: string | null;
  status: ProductStatus;
  price: number;
  cost_price: number | null;
  discount: number;
  currency: string;
  stock_qty: number;
  low_stock_threshold: number | null;
  images: string[];
  created_at: string;
  updated_at: string;
};

export type CreateProductInput = {
  name: string;
  description?: string;
  sku?: string;
  status?: ProductStatus;
  price: number;
  cost_price?: number;
  discount?: number;
  currency?: string;
  stock_qty?: number;
  low_stock_threshold?: number;
  images?: string[];
};

// Single update type — all fields optional, nullable fields explicit
export type UpdateProductInput = {
  name?: string;
  description?: string | null;
  sku?: string | null;
  status?: ProductStatus;
  price?: number;
  cost_price?: number | null;
  discount?: number | null;
  currency?: string;
  stock_qty?: number;
  low_stock_threshold?: number | null;
  images?: string[];
};

export type ListProductsParams = {
  status?: ProductStatus;
  search?: string;
  limit?: number;
  offset?: number;
};
