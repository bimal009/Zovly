"use client";

import * as React from "react";
import {
  useForm,
  useFieldArray,
  Controller,
  type Control,
  type UseFormRegister,
  type FieldErrors,
  type Resolver,
} from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Trash2, X } from "lucide-react";
import { Badge } from "@repo/ui/components/ui/badge";
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
import { useGetCategories } from "@/features/categories/client/categories";
import {
  ProductStatusSchema,
  type CreateProductInput,
  type CreateProductVariantInput,
  type Product,
  type ProductVariant,
  type UpdateProductInput,
} from "../types/products";

// ─── Form schema ──────────────────────────────────────────────────────────────

// z.coerce.number() handles the string→number coercion that RHF feeds in,
// so the inferred type stays `number | undefined` — no transform mismatch.
const optionalNonNeg = z.coerce.number().nonnegative().optional();
const optionalNonNegInt = z.coerce.number().int().nonnegative().optional();

// A variant option, e.g. { key: "size", values: ["S", "M", "L"] }
const VariantAttributeSchema = z.object({
  key: z.string(),
  values: z.array(z.string()),
});

const VariantSchema = z.object({
  name: z.string().min(1, "Variant name is required"),
  sku: z.string(),
  // 0/empty means inherit the parent product value
  price: z.coerce.number().min(0),
  cost_price: optionalNonNeg,
  discount: z.coerce.number().int().min(0).max(100),
  stock_qty: z.coerce.number().int().min(0),
  low_stock_threshold: optionalNonNegInt,
  attributes: z.array(VariantAttributeSchema),
});

// "" means "no category"
const NO_CATEGORY = "";

const ProductFormSchema = z.object({
  category_id: z.string(),
  name: z.string().min(1, "Name is required"),
  description: z.string().max(200, "Must be 200 characters or less"),
  sku: z.string(),
  status: ProductStatusSchema,
  price: z.coerce.number().positive("Must be > 0"),
  cost_price: optionalNonNeg,
  discount: z.coerce.number().int().min(0).max(100),
  currency: z.string(),
  stock_qty: z.coerce.number().int().min(0),
  low_stock_threshold: optionalNonNegInt,
  images: z.array(z.string()),
  variants: z.array(VariantSchema),
});

type ProductFormValues = z.infer<typeof ProductFormSchema>;
type VariantFormValues = z.infer<typeof VariantSchema>;

const EMPTY_VARIANT: VariantFormValues = {
  name: "",
  sku: "",
  price: 0,
  cost_price: undefined,
  discount: 0,
  stock_qty: 0,
  low_stock_threshold: undefined,
  attributes: [],
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toCents(val: number) {
  return Math.round(val * 100);
}

// Maps a saved variant back into editable form values.
function variantToFormValues(v: ProductVariant): VariantFormValues {
  const attributes = Object.entries(v.attributes ?? {}).map(([key, value]) => ({
    key,
    // values were stored comma-joined ("S, M, L") — split them back into tags
    values: value
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  }));

  return {
    name: v.name,
    sku: v.sku ?? "",
    price: v.price != null ? v.price / 100 : 0,
    cost_price: v.cost_price != null ? v.cost_price / 100 : undefined,
    discount: v.discount ?? 0,
    stock_qty: v.stock_qty,
    low_stock_threshold: v.low_stock_threshold ?? undefined,
    attributes,
  };
}

const DEFAULT_VALUES: ProductFormValues = {
  category_id: NO_CATEGORY,
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
  variants: [],
};

function toFormValues(p: Product): ProductFormValues {
  return {
    category_id: p.category_id ?? NO_CATEGORY,
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
    variants: (p.variants ?? []).map(variantToFormValues),
  };
}

function buildVariantInput(v: VariantFormValues): CreateProductVariantInput {
  const attributes = v.attributes.reduce<Record<string, string>>((acc, a) => {
    const key = a.key.trim();
    const values = a.values.map((s) => s.trim()).filter(Boolean);
    if (key && values.length) acc[key] = values.join(", ");
    return acc;
  }, {});

  return {
    name: v.name.trim(),
    sku: v.sku.trim() || undefined,
    price: v.price > 0 ? toCents(v.price) : undefined,
    cost_price: v.cost_price && v.cost_price > 0 ? toCents(v.cost_price) : undefined,
    discount: v.discount > 0 ? v.discount : undefined,
    stock_qty: v.stock_qty,
    low_stock_threshold: v.low_stock_threshold,
    attributes: Object.keys(attributes).length ? attributes : undefined,
  };
}

function buildCreateInput(v: ProductFormValues): CreateProductInput {
  return {
    category_id: v.category_id || undefined,
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
    variants: v.variants.length
      ? v.variants.map(buildVariantInput)
      : undefined,
  };
}

function buildUpdateInput(v: ProductFormValues): UpdateProductInput {
  return {
    category_id: v.category_id || null,
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
    // sent as a full replacement set — backend wiring is pending
    variants: v.variants.map(buildVariantInput),
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

  const {
    fields: variantFields,
    append: appendVariant,
    remove: removeVariant,
  } = useFieldArray({ control, name: "variants" });

  const { data: categoriesData } = useGetCategories();
  const categories = categoriesData?.data ?? [];

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
                    maxLength={200}
                    {...register("description")}
                  />
                  {errors.description && (
                    <p className="text-xs text-destructive">
                      {errors.description.message}
                    </p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5 sm:col-span-2">
                  <Label htmlFor="p-category">Category</Label>
                  <Controller
                    name="category_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        value={field.value || "none"}
                        onValueChange={(v) =>
                          field.onChange(v === "none" ? "" : v)
                        }
                      >
                        <SelectTrigger id="p-category" className="w-full">
                          <SelectValue placeholder="Uncategorized" />
                        </SelectTrigger>
                        <SelectContent className="p-2">
                          <SelectItem value="none">Uncategorized</SelectItem>
                          {categories.map((c) => (
                            <SelectItem
                              key={c.id}
                              value={c.id}
                              className="capitalize"
                            >
                              {c.name.toLowerCase()}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
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
                        <SelectTrigger id="p-status" className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent className="p-2">
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
                        <SelectTrigger id="p-currency" className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent className="p-2">
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

            <Separator />

            {/* Variants */}
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-semibold">Variants</p>
                  <p className="text-xs text-muted-foreground">
                    Optional. Sizes, colors or options with their own stock and
                    pricing. Leave price empty to inherit the product price.
                  </p>
                </div>
              </div>

              {variantFields.length > 0 && (
                <div className="flex flex-col gap-3">
                  {variantFields.map((field, index) => (
                    <VariantRow
                      key={field.id}
                      index={index}
                      control={control}
                      register={register}
                      errors={errors}
                      onRemove={() => removeVariant(index)}
                    />
                  ))}
                </div>
              )}

              <Button
                type="button"
                variant="outline"
                size="sm"
                className="w-fit cursor-pointer"
                onClick={() => appendVariant({ ...EMPTY_VARIANT })}
              >
                <Plus className="size-4" aria-hidden="true" />
                Add Variant
              </Button>
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

// ─── Tag input (multi-value option) ───────────────────────────────────────────

interface TagInputProps {
  value: string[];
  onChange: (next: string[]) => void;
  placeholder?: string;
  ariaLabel?: string;
}

function TagInput({ value, onChange, placeholder, ariaLabel }: TagInputProps) {
  const [draft, setDraft] = React.useState("");

  function commit() {
    const parts = draft
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (parts.length) {
      const next = [...value];
      for (const p of parts) {
        if (!next.includes(p)) next.push(p);
      }
      onChange(next);
    }
    setDraft("");
  }

  return (
    <div className="flex min-h-9 flex-1 flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1.5 focus-within:ring-1 focus-within:ring-ring">
      {value.map((tag, i) => (
        <Badge key={`${tag}-${i}`} variant="secondary" className="gap-1 pr-1">
          {tag}
          <button
            type="button"
            className="cursor-pointer rounded-sm opacity-70 hover:opacity-100"
            onClick={() => onChange(value.filter((_, idx) => idx !== i))}
            aria-label={`Remove ${tag}`}
          >
            <X className="size-3" aria-hidden="true" />
          </button>
        </Badge>
      ))}
      <input
        value={draft}
        aria-label={ariaLabel}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === ",") {
            e.preventDefault();
            commit();
          } else if (e.key === "Backspace" && !draft && value.length) {
            onChange(value.slice(0, -1));
          }
        }}
        onBlur={commit}
        placeholder={value.length ? "" : placeholder}
        className="min-w-[100px] flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
      />
    </div>
  );
}

// ─── Variant row (create mode) ─────────────────────────────────────────────────

interface VariantRowProps {
  index: number;
  control: Control<ProductFormValues>;
  register: UseFormRegister<ProductFormValues>;
  errors: FieldErrors<ProductFormValues>;
  onRemove: () => void;
}

function VariantRow({
  index,
  control,
  register,
  errors,
  onRemove,
}: VariantRowProps) {
  const {
    fields: attrFields,
    append: appendAttr,
    remove: removeAttr,
  } = useFieldArray({ control, name: `variants.${index}.attributes` });

  const variantErrors = errors.variants?.[index];

  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-4">
      <div className="flex items-start justify-between gap-2">
        <p className="text-sm font-medium">Variant {index + 1}</p>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="size-7 cursor-pointer text-muted-foreground hover:text-destructive"
          onClick={onRemove}
          aria-label={`Remove variant ${index + 1}`}
        >
          <Trash2 className="size-4" aria-hidden="true" />
        </Button>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="flex flex-col gap-1.5 sm:col-span-2">
          <Label htmlFor={`v-${index}-name`}>Variant Name *</Label>
          <Input
            id={`v-${index}-name`}
            placeholder="e.g. Red / Medium"
            {...register(`variants.${index}.name`)}
          />
          {variantErrors?.name && (
            <p className="text-xs text-destructive">
              {variantErrors.name.message}
            </p>
          )}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-sku`}>SKU</Label>
          <Input
            id={`v-${index}-sku`}
            placeholder="e.g. BJ-RED-M"
            {...register(`variants.${index}.sku`)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-stock`}>Stock Qty</Label>
          <Input
            id={`v-${index}-stock`}
            type="number"
            min="0"
            placeholder="0"
            {...register(`variants.${index}.stock_qty`)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-price`}>Selling Price</Label>
          <Input
            id={`v-${index}-price`}
            type="number"
            min="0"
            step="0.01"
            placeholder="Inherit product price"
            {...register(`variants.${index}.price`)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-cost`}>Cost Price</Label>
          <Input
            id={`v-${index}-cost`}
            type="number"
            min="0"
            step="0.01"
            placeholder="0.00"
            {...register(`variants.${index}.cost_price`)}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-discount`}>Discount (%)</Label>
          <Input
            id={`v-${index}-discount`}
            type="number"
            min="0"
            max="100"
            placeholder="0"
            {...register(`variants.${index}.discount`)}
          />
          {variantErrors?.discount && (
            <p className="text-xs text-destructive">
              {variantErrors.discount.message}
            </p>
          )}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={`v-${index}-threshold`}>Low Stock Alert at</Label>
          <Input
            id={`v-${index}-threshold`}
            type="number"
            min="0"
            placeholder="e.g. 5"
            {...register(`variants.${index}.low_stock_threshold`)}
          />
        </div>
      </div>

      {/* Options / attributes */}
      <div className="flex flex-col gap-2">
        <Label className="text-xs text-muted-foreground">
          Options (e.g. size → S, M, L)
        </Label>
        {attrFields.map((attr, ai) => (
          <div
            key={attr.id}
            className="flex flex-col gap-2 sm:flex-row sm:items-start"
          >
            <Input
              placeholder="Option name (e.g. size)"
              aria-label={`Option name ${ai + 1}`}
              className="sm:w-40 sm:shrink-0"
              {...register(`variants.${index}.attributes.${ai}.key`)}
            />
            <div className="flex flex-1 items-start gap-2">
              <Controller
                control={control}
                name={`variants.${index}.attributes.${ai}.values`}
                render={({ field }) => (
                  <TagInput
                    value={field.value}
                    onChange={field.onChange}
                    placeholder="Type a value, press Enter or comma"
                    ariaLabel={`Option values ${ai + 1}`}
                  />
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-9 shrink-0 cursor-pointer text-muted-foreground hover:text-destructive"
                onClick={() => removeAttr(ai)}
                aria-label={`Remove option ${ai + 1}`}
              >
                <Trash2 className="size-4" aria-hidden="true" />
              </Button>
            </div>
          </div>
        ))}
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="w-fit cursor-pointer text-muted-foreground"
          onClick={() => appendAttr({ key: "", values: [] })}
        >
          <Plus className="size-4" aria-hidden="true" />
          Add Option
        </Button>
      </div>
    </div>
  );
}
