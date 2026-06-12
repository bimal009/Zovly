"use client";

import * as React from "react";
import { useForm, Controller, type Resolver } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@repo/ui/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@repo/ui/components/ui/dialog";
import { Input } from "@repo/ui/components/ui/input";
import { Label } from "@repo/ui/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@repo/ui/components/ui/select";
import { Separator } from "@repo/ui/components/ui/separator";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { ImageUploader } from "@/components/ImageUploader";
import {
  ProductStatusSchema,
  type CreateProductInput,
  type Product,
  type UpdateProductInput,
} from "../types/products";

// ─── Form schema ──────────────────────────────────────────────────────────────

// z.coerce.number() handles the string→number coercion that RHF feeds in,
// so the inferred type stays `number | undefined` — no transform mismatch.
const optionalNonNeg = z.coerce.number().nonnegative().optional();
const optionalNonNegInt = z.coerce.number().int().nonnegative().optional();

const ProductFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string(),
  sku: z.string(),
  status: ProductStatusSchema,
  price: z.coerce.number().positive("Must be > 0"),
  cost_price: optionalNonNeg,
  discount: z.coerce.number().int().min(0).max(100),
  currency: z.string(),
  stock_qty: z.coerce.number().int().min(0),
  low_stock_threshold: optionalNonNegInt,
  images: z.array(z.string()),
});

type ProductFormValues = z.infer<typeof ProductFormSchema>;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toCents(val: number) {
  return Math.round(val * 100);
}

const DEFAULT_VALUES: ProductFormValues = {
  name: "",
  description: "",
  sku: "",
  status: "active",
  price: 0,
  cost_price: undefined,
  discount: 0,
  currency: "USD",
  stock_qty: 0,
  low_stock_threshold: undefined,
  images: [],
};

function toFormValues(p: Product): ProductFormValues {
  return {
    name: p.name,
    description: p.description ?? "",
    sku: p.sku ?? "",
    status: p.status,
    price: p.price / 100,
    cost_price: p.cost_price != null ? p.cost_price / 100 : undefined,
    discount: p.discount ?? 0,
    currency: p.currency,
    stock_qty: p.stock_qty,
    low_stock_threshold: p.low_stock_threshold ?? undefined,
    images: p.images ?? [],
  };
}

function buildCreateInput(v: ProductFormValues): CreateProductInput {
  return {
    name: v.name.trim(),
    description: v.description.trim() || undefined,
    sku: v.sku.trim() || undefined,
    status: v.status,
    price: toCents(v.price),
    cost_price: v.cost_price != null ? toCents(v.cost_price) : undefined,
    discount: v.discount,
    currency: v.currency,
    stock_qty: v.stock_qty,
    low_stock_threshold: v.low_stock_threshold,
    images: v.images.length ? v.images : undefined,
  };
}

function buildUpdateInput(v: ProductFormValues): UpdateProductInput {
  return {
    name: v.name.trim(),
    description: v.description.trim() || null,
    sku: v.sku.trim() || null,
    status: v.status,
    price: toCents(v.price),
    cost_price: v.cost_price != null ? toCents(v.cost_price) : null,
    discount: v.discount,
    currency: v.currency,
    stock_qty: v.stock_qty,
    low_stock_threshold: v.low_stock_threshold ?? null,
    images: v.images,
  };
}

// ─── Component ────────────────────────────────────────────────────────────────

interface ProductFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: Product | null;
  onSave: (
    data: CreateProductInput | { id: string; input: UpdateProductInput },
  ) => void;
  saving: boolean;
}

export function ProductFormDialog({
  open,
  onOpenChange,
  editing,
  onSave,
  saving,
}: ProductFormDialogProps) {
  const {
    register,
    handleSubmit,
    control,
    reset,
    setValue,
    formState: { errors },
  } = useForm<ProductFormValues>({
    resolver: zodResolver(ProductFormSchema) as Resolver<ProductFormValues>,
    defaultValues: DEFAULT_VALUES,
  });

  React.useEffect(() => {
    reset(editing ? toFormValues(editing) : DEFAULT_VALUES);
  }, [editing, open, reset]);

  const handleImagesChange = React.useCallback(
    (urls: string[]) => setValue("images", urls, { shouldDirty: true }),
    [setValue],
  );

  function onSubmit(values: ProductFormValues) {
    if (editing) {
      onSave({ id: editing.id, input: buildUpdateInput(values) });
    } else {
      onSave(buildCreateInput(values));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex flex-col max-w-3xl sm:max-w-3xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit Product" : "Add Product"}</DialogTitle>
        </DialogHeader>

        <form
          id="product-form"
          onSubmit={handleSubmit(onSubmit)}
          className="flex-1 overflow-y-auto py-2 pr-1"
        >
          <div className="flex flex-col gap-5">
            {/* Basic */}
            <div className="flex flex-col gap-3">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="flex flex-col gap-1.5 sm:col-span-2">
                  <Label htmlFor="p-name">Name *</Label>
                  <Input
                    id="p-name"
                    placeholder="e.g. Blue Jacket (M)"
                    {...register("name")}
                  />
                  {errors.name && (
                    <p className="text-xs text-destructive">
                      {errors.name.message}
                    </p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5 sm:col-span-2">
                  <Label htmlFor="p-desc">Description</Label>
                  <Textarea
                    id="p-desc"
                    placeholder="Short product description…"
                    rows={2}
                    {...register("description")}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-sku">SKU</Label>
                  <Input
                    id="p-sku"
                    placeholder="e.g. BJ-M-001"
                    {...register("sku")}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-status">Status</Label>
                  <Controller
                    name="status"
                    control={control}
                    render={({ field }) => (
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <SelectTrigger id="p-status">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="active">Active</SelectItem>
                          <SelectItem value="inactive">Inactive</SelectItem>
                          <SelectItem value="archived">Archived</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* Pricing */}
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold">Pricing</p>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-price">Selling Price *</Label>
                  <Input
                    id="p-price"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    {...register("price")}
                  />
                  {errors.price && (
                    <p className="text-xs text-destructive">
                      {errors.price.message}
                    </p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-cost">Cost Price</Label>
                  <Input
                    id="p-cost"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    {...register("cost_price")}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-discount">Discount (%)</Label>
                  <Input
                    id="p-discount"
                    type="number"
                    min="0"
                    max="100"
                    placeholder="0"
                    {...register("discount")}
                  />
                  {errors.discount && (
                    <p className="text-xs text-destructive">
                      {errors.discount.message}
                    </p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-currency">Currency</Label>
                  <Controller
                    name="currency"
                    control={control}
                    render={({ field }) => (
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <SelectTrigger id="p-currency">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="USD">USD</SelectItem>
                          <SelectItem value="EUR">EUR</SelectItem>
                          <SelectItem value="GBP">GBP</SelectItem>
                          <SelectItem value="INR">INR</SelectItem>
                          <SelectItem value="NPR">NPR</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* Inventory */}
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold">Inventory</p>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-stock">Stock Qty</Label>
                  <Input
                    id="p-stock"
                    type="number"
                    min="0"
                    placeholder="0"
                    {...register("stock_qty")}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-threshold">Low Stock Alert at</Label>
                  <Input
                    id="p-threshold"
                    type="number"
                    min="0"
                    placeholder="e.g. 5"
                    {...register("low_stock_threshold")}
                  />
                </div>
              </div>
            </div>

            <Separator />

            {/* Images */}
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold">Images</p>
              <ImageUploader
                key={editing?.id ?? "new"}
                folder="/products"
                initialUrls={editing?.images ?? []}
                onUploadsChange={handleImagesChange}
              />
            </div>
          </div>
        </form>

        <DialogFooter>
          <Button
            variant="outline"
            className="cursor-pointer"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="product-form"
            className="cursor-pointer"
            disabled={saving}
          >
            {saving ? "Saving…" : editing ? "Save Changes" : "Add Product"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
