import { z } from "zod";

export const ProductStatusSchema = z.enum(["active", "inactive", "archived"]);
export type ProductStatus = z.infer<typeof ProductStatusSchema>;

export type ProductVariant = {
  id: string;
  product_id: string;
  business_id: string;
  name: string;
  sku: string | null;
  // structured option values — { "color": "red", "size": "M" }
  attributes?: Record<string, string> | null;
  // null pricing/threshold means inherit the parent product's value
  price: number | null;
  cost_price: number | null;
  discount: number | null;
  stock_qty: number;
  low_stock_threshold: number | null;
  images: string[];
  created_at: string;
  updated_at: string;
};

export type CreateProductVariantInput = {
  name: string;
  sku?: string;
  attributes?: Record<string, string>;
  price?: number; // selling price; omit to inherit parent product price
  cost_price?: number;
  discount?: number;
  stock_qty?: number;
  low_stock_threshold?: number;
  images?: string[];
};

export type Product = {
  id: string;
  business_id: string;
  category_id: string | null;
  name: string;
  description: string | null;
  sku: string | null;
  status: ProductStatus;
  tags: string[];
  // structured product-level attributes — { "material": "cotton", "fit": "slim" }
  attributes?: Record<string, string> | null;
  price: number;
  cost_price: number | null;
  discount: number;
  currency: string;
  stock_qty: number;
  low_stock_threshold: number | null;
  images: string[];
  variants?: ProductVariant[];
  created_at: string;
  updated_at: string;
};

export type CreateProductInput = {
  category_id?: string;
  name: string;
  description?: string;
  sku?: string;
  status?: ProductStatus;
  tags?: string[];
  attributes?: Record<string, string>;
  price: number;
  cost_price?: number;
  discount?: number;
  currency?: string;
  stock_qty?: number;
  low_stock_threshold?: number;
  images?: string[];
  variants?: CreateProductVariantInput[];
};

// Single update type — all fields optional, nullable fields explicit
export type UpdateProductInput = {
  category_id?: string | null;
  name?: string;
  description?: string | null;
  sku?: string | null;
  status?: ProductStatus;
  tags?: string[];
  attributes?: Record<string, string> | null;
  price?: number;
  cost_price?: number | null;
  discount?: number | null;
  currency?: string;
  stock_qty?: number;
  low_stock_threshold?: number | null;
  images?: string[];
  variants?: CreateProductVariantInput[];
};

export type ListProductsParams = {
  status?: ProductStatus;
  search?: string;
  limit?: number;
  offset?: number;
};
