"use client";

import * as React from "react";
import { useForm, Controller, useFieldArray, type Resolver } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@repo/ui/components/ui/button";
import { Plus, Trash2 } from "lucide-react";
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
import { Switch } from "@repo/ui/components/ui/switch";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { ImageUploader } from "@/components/ImageUploader";
import type {
  CreateServiceInput,
  Service,
  ServiceFeature,
  UpdateServiceInput,
} from "../api/services";

// ─── Schema ────────────────────────────────────────────────────────────────────

const emptyToUndef = (v: unknown) =>
  v === "" || v === null || v === undefined ? undefined : v;

const optPosInt = z.preprocess(
  emptyToUndef,
  z.coerce.number().int().positive().optional(),
);
const optNonNegInt = z.preprocess(
  emptyToUndef,
  z.coerce.number().int().nonnegative().optional(),
);
const optNonNeg = z.preprocess(
  emptyToUndef,
  z.coerce.number().nonnegative().optional(),
);

const ServiceFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  type: z.enum(["appointment", "membership", "class", "package"]),
  status: z.enum(["active", "inactive", "archived"]),
  description: z.string(),
  price: z.coerce.number().nonnegative("Must be ≥ 0"),
  currency: z.string(),
  requires_deposit: z.boolean(),
  deposit_amount: optNonNeg,
  duration_min: optPosInt,
  buffer_min: optNonNegInt,
  max_advance_days: optPosInt,
  max_concurrent: optPosInt,
  billing_interval: z
    .enum(["weekly", "monthly", "quarterly", "yearly"])
    .optional(),
  trial_days: optNonNegInt,
  session_count: optPosInt,
  validity_days: optPosInt,
  features: z.array(
    z.object({
      label: z.string().min(1, "Required"),
      value: z.string().min(1, "Required"),
    }),
  ),
  images: z.array(z.string()),
});

type ServiceFormValues = z.infer<typeof ServiceFormSchema>;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function toCents(val: number) {
  return Math.round(val * 100);
}

const DEFAULT_VALUES: ServiceFormValues = {
  name: "",
  type: "appointment",
  status: "active",
  description: "",
  price: 0,
  currency: "USD",
  requires_deposit: false,
  deposit_amount: undefined,
  duration_min: 60,
  buffer_min: 15,
  max_advance_days: 30,
  max_concurrent: undefined,
  billing_interval: "monthly",
  trial_days: undefined,
  session_count: undefined,
  validity_days: undefined,
  features: [],
  images: [],
};

function toFormValues(s: Service): ServiceFormValues {
  return {
    name: s.name,
    type: s.type,
    status: s.status,
    description: s.description ?? "",
    price: s.price / 100,
    currency: s.currency,
    requires_deposit: s.requires_deposit,
    deposit_amount:
      s.deposit_amount != null ? s.deposit_amount / 100 : undefined,
    duration_min: s.duration_min ?? undefined,
    buffer_min: s.buffer_min ?? undefined,
    max_advance_days: s.max_advance_days ?? undefined,
    max_concurrent: s.max_concurrent ?? undefined,
    billing_interval: s.billing_interval ?? undefined,
    trial_days: s.trial_days ?? undefined,
    session_count: s.session_count ?? undefined,
    validity_days: s.validity_days ?? undefined,
    features: s.features ?? [],
    images: s.images ?? [],
  };
}

function buildCreateInput(v: ServiceFormValues): CreateServiceInput {
  const hasDuration =
    v.type === "appointment" || v.type === "class" || v.type === "package";
  return {
    name: v.name.trim(),
    type: v.type,
    status: v.status,
    description: v.description.trim() || undefined,
    price: toCents(v.price),
    currency: v.currency,
    requires_deposit: v.requires_deposit,
    deposit_amount:
      v.requires_deposit && v.deposit_amount != null
        ? toCents(v.deposit_amount)
        : undefined,
    duration_min: hasDuration ? v.duration_min : undefined,
    buffer_min: hasDuration ? v.buffer_min : undefined,
    max_advance_days: hasDuration ? v.max_advance_days : undefined,
    max_concurrent: v.type === "class" ? v.max_concurrent : undefined,
    billing_interval:
      v.type === "membership" ? v.billing_interval : undefined,
    trial_days: v.type === "membership" ? v.trial_days : undefined,
    session_count: v.type === "package" ? v.session_count : undefined,
    validity_days: v.type === "package" ? v.validity_days : undefined,
    features: v.features.length ? v.features : undefined,
    images: v.images.length ? v.images : undefined,
  };
}

function buildUpdateInput(v: ServiceFormValues): UpdateServiceInput {
  const hasDuration =
    v.type === "appointment" || v.type === "class" || v.type === "package";
  return {
    name: v.name.trim(),
    status: v.status,
    description: v.description.trim() || null,
    price: toCents(v.price),
    currency: v.currency,
    requires_deposit: v.requires_deposit,
    deposit_amount:
      v.requires_deposit && v.deposit_amount != null
        ? toCents(v.deposit_amount)
        : null,
    duration_min: hasDuration ? (v.duration_min ?? null) : null,
    buffer_min: hasDuration ? (v.buffer_min ?? null) : null,
    max_advance_days: hasDuration ? (v.max_advance_days ?? null) : null,
    max_concurrent: v.type === "class" ? (v.max_concurrent ?? null) : null,
    billing_interval:
      v.type === "membership" ? (v.billing_interval ?? null) : null,
    trial_days: v.type === "membership" ? (v.trial_days ?? null) : null,
    session_count: v.type === "package" ? (v.session_count ?? null) : null,
    validity_days: v.type === "package" ? (v.validity_days ?? null) : null,
    features: v.features as ServiceFeature[],
    images: v.images,
  };
}

// ─── Component ────────────────────────────────────────────────────────────────

interface ServiceFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: Service | null;
  onSave: (
    data: CreateServiceInput | { id: string; input: UpdateServiceInput },
  ) => void;
  saving: boolean;
}

export function ServiceFormDialog({
  open,
  onOpenChange,
  editing,
  onSave,
  saving,
}: ServiceFormDialogProps) {
  const {
    register,
    handleSubmit,
    control,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ServiceFormValues>({
    resolver: zodResolver(ServiceFormSchema) as Resolver<ServiceFormValues>,
    defaultValues: DEFAULT_VALUES,
  });

  React.useEffect(() => {
    reset(editing ? toFormValues(editing) : DEFAULT_VALUES);
  }, [editing, open, reset]);

  const { fields: featureFields, append: appendFeature, remove: removeFeature } = useFieldArray({
    control,
    name: "features",
  });

  const serviceType = watch("type");
  const requiresDeposit = watch("requires_deposit");
  const hasDuration =
    serviceType === "appointment" ||
    serviceType === "class" ||
    serviceType === "package";

  const handleImagesChange = React.useCallback(
    (urls: string[]) => setValue("images", urls, { shouldDirty: true }),
    [setValue],
  );

  function onSubmit(values: ServiceFormValues) {
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
          <DialogTitle>
            {editing ? "Edit Service" : "Add Service"}
          </DialogTitle>
        </DialogHeader>

        <form
          id="service-form"
          onSubmit={handleSubmit(onSubmit)}
          className="flex-1 overflow-y-auto py-2 pr-1"
        >
          <div className="flex flex-col gap-5">
            {/* Core */}
            <div className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="s-name">Name *</Label>
                <Input
                  id="s-name"
                  placeholder="e.g. 1-on-1 Business Consultation"
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-xs text-destructive">
                    {errors.name.message}
                  </p>
                )}
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="s-type">Type *</Label>
                  <Controller
                    name="type"
                    control={control}
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger id="s-type">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="appointment">Appointment</SelectItem>
                          <SelectItem value="class">Class</SelectItem>
                          <SelectItem value="membership">Membership</SelectItem>
                          <SelectItem value="package">Package</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="s-status">Status</Label>
                  <Controller
                    name="status"
                    control={control}
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger id="s-status">
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
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="s-desc">Description</Label>
                <Textarea
                  id="s-desc"
                  placeholder="What does this service include?"
                  rows={2}
                  {...register("description")}
                />
              </div>
            </div>

            <Separator />

            {/* Pricing */}
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold">Pricing</p>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="s-price">Price *</Label>
                  <Input
                    id="s-price"
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
                  <Label htmlFor="s-currency">Currency</Label>
                  <Controller
                    name="currency"
                    control={control}
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange}>
                        <SelectTrigger id="s-currency">
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
              <div className="flex items-center justify-between rounded-lg border p-3">
                <div>
                  <p className="text-sm font-medium">Require Deposit</p>
                  <p className="text-xs text-muted-foreground">
                    Collect upfront to secure bookings
                  </p>
                </div>
                <Controller
                  name="requires_deposit"
                  control={control}
                  render={({ field }) => (
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  )}
                />
              </div>
              {requiresDeposit && (
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="s-deposit">Deposit Amount</Label>
                  <Input
                    id="s-deposit"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    {...register("deposit_amount")}
                  />
                </div>
              )}
            </div>

            <Separator />

            {/* Scheduling (appointment / class / package) */}
            {hasDuration && (
              <div className="flex flex-col gap-3">
                <p className="text-sm font-semibold">Scheduling</p>
                <div className="grid grid-cols-3 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-duration">Duration (min)</Label>
                    <Input
                      id="s-duration"
                      type="number"
                      min="1"
                      placeholder="60"
                      {...register("duration_min")}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-buffer">Buffer (min)</Label>
                    <Input
                      id="s-buffer"
                      type="number"
                      min="0"
                      placeholder="15"
                      {...register("buffer_min")}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-advance">Book Ahead (days)</Label>
                    <Input
                      id="s-advance"
                      type="number"
                      min="1"
                      placeholder="30"
                      {...register("max_advance_days")}
                    />
                  </div>
                </div>
                {serviceType === "class" && (
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-concurrent">Max Participants</Label>
                    <Input
                      id="s-concurrent"
                      type="number"
                      min="1"
                      placeholder="e.g. 20"
                      {...register("max_concurrent")}
                    />
                  </div>
                )}
              </div>
            )}

            {/* Membership */}
            {serviceType === "membership" && (
              <div className="flex flex-col gap-3">
                <p className="text-sm font-semibold">Membership</p>
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-billing">Billing Interval</Label>
                    <Controller
                      name="billing_interval"
                      control={control}
                      render={({ field }) => (
                        <Select
                          value={field.value ?? ""}
                          onValueChange={(v) =>
                            field.onChange(v || undefined)
                          }
                        >
                          <SelectTrigger id="s-billing">
                            <SelectValue placeholder="Select interval" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="weekly">Weekly</SelectItem>
                            <SelectItem value="monthly">Monthly</SelectItem>
                            <SelectItem value="quarterly">Quarterly</SelectItem>
                            <SelectItem value="yearly">Yearly</SelectItem>
                          </SelectContent>
                        </Select>
                      )}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-trial">Trial Days</Label>
                    <Input
                      id="s-trial"
                      type="number"
                      min="0"
                      placeholder="e.g. 7"
                      {...register("trial_days")}
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Package */}
            {serviceType === "package" && (
              <div className="flex flex-col gap-3">
                <p className="text-sm font-semibold">Package Details</p>
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-sessions">Sessions</Label>
                    <Input
                      id="s-sessions"
                      type="number"
                      min="1"
                      placeholder="e.g. 5"
                      {...register("session_count")}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <Label htmlFor="s-validity">Validity (days)</Label>
                    <Input
                      id="s-validity"
                      type="number"
                      min="1"
                      placeholder="e.g. 90"
                      {...register("validity_days")}
                    />
                  </div>
                </div>
              </div>
            )}

            <Separator />

            {/* Features / inclusions */}
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-semibold">What's Included</p>
                  <p className="text-xs text-muted-foreground">
                    Benefits or inclusions shown to customers
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => appendFeature({ label: "", value: "" })}
                >
                  <Plus className="mr-1.5 h-3.5 w-3.5" />
                  Add
                </Button>
              </div>
              {featureFields.length > 0 && (
                <div className="flex flex-col gap-2">
                  {featureFields.map((field, index) => (
                    <div key={field.id} className="flex items-start gap-2">
                      <div className="flex-1">
                        <Input
                          placeholder="Feature (e.g. PT Sessions)"
                          {...register(`features.${index}.label`)}
                        />
                        {errors.features?.[index]?.label && (
                          <p className="mt-0.5 text-xs text-destructive">
                            {errors.features[index]?.label?.message}
                          </p>
                        )}
                      </div>
                      <div className="flex-1">
                        <Input
                          placeholder="Value (e.g. 2/month)"
                          {...register(`features.${index}.value`)}
                        />
                        {errors.features?.[index]?.value && (
                          <p className="mt-0.5 text-xs text-destructive">
                            {errors.features[index]?.value?.message}
                          </p>
                        )}
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="mt-0.5 h-9 w-9 shrink-0 cursor-pointer text-muted-foreground hover:text-destructive"
                        onClick={() => removeFeature(index)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <Separator />

            {/* Images */}
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold">Images</p>
              <ImageUploader
                key={editing?.id ?? "new"}
                folder="/services"
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
            form="service-form"
            className="cursor-pointer"
            disabled={saving}
          >
            {saving ? "Saving…" : editing ? "Save Changes" : "Add Service"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
