"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { isAxiosError } from "axios";
import {
  Controller,
  useForm,
  type FieldErrors,
  type Resolver,
} from "react-hook-form";
import * as z from "zod";
import {
  AlertCircle,
  Check,
  LayoutGrid,
  ShoppingBag,
  Wrench,
} from "lucide-react";
import { createBusiness } from "@/features/onboard/api/business";
import { ImageUploader } from "@/components/ImageUploader";
import { CreatingDashboard } from "./CreatingDashboard";
import { Button } from "@repo/ui/components/ui/button";
import { Card } from "@repo/ui/components/ui/card";
import { CountryDropdown } from "@repo/ui/components/ui/country-dropdown";
import { Input } from "@repo/ui/components/ui/input";
import { Textarea } from "@repo/ui/components/ui/textarea";
import { cn } from "@repo/ui/utils";
import { createBusinessSchema } from "@repo/types";
import type { CreateBusinessInput } from "@repo/types";

type OnboardingFormValues = z.input<typeof createBusinessSchema>;

const ONBOARDING_FIELDS = [
  "name",
  "description",
  "logo",
  "website",
  "phone",
  "address",
  "city",
  "country",
  "type",
] as const satisfies readonly (keyof OnboardingFormValues)[];

type OnboardingField = (typeof ONBOARDING_FIELDS)[number];

const BUSINESS_TYPES = [
  {
    value: "product" as const,
    label: "Product seller",
    description: "You sell physical or digital products",
    icon: ShoppingBag,
  },
  {
    value: "service" as const,
    label: "Service business",
    description: "You sell services",
    icon: Wrench,
  },
  {
    value: "both" as const,
    label: "Both",
    description: "You sell products and services",
    icon: LayoutGrid,
  },
];

type Phase = "form" | "creating" | "ready";

const isOnboardingField = (field: PropertyKey): field is OnboardingField =>
  typeof field === "string" &&
  ONBOARDING_FIELDS.includes(field as OnboardingField);

const optionalUrlValue = (value: string) => {
  const trimmedValue = value.trim();
  return trimmedValue.length > 0 ? trimmedValue : undefined;
};

const toFieldErrors = (
  issues: z.ZodIssue[],
): FieldErrors<OnboardingFormValues> => {
  const errors: FieldErrors<OnboardingFormValues> = {};

  for (const issue of issues) {
    const fieldName = issue.path[0];

    if (fieldName === undefined || !isOnboardingField(fieldName)) {
      continue;
    }

    errors[fieldName] ??= {
      type: issue.code,
      message: issue.message,
    };
  }

  return errors;
};

const createBusinessResolver: Resolver<
  OnboardingFormValues,
  undefined,
  CreateBusinessInput
> = async (values) => {
  const result = await createBusinessSchema.safeParseAsync(values);

  if (result.success) {
    return {
      values: result.data,
      errors: {},
    };
  }

  return {
    values: {},
    errors: toFieldErrors(result.error.issues),
  };
};

function FieldError({ message }: { message?: string }) {
  if (!message) return null;

  return (
    <p
      role="alert"
      className="flex items-center gap-1.5 text-xs font-medium text-destructive motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-top-1 duration-200"
    >
      <AlertCircle className="w-3 h-3 shrink-0" aria-hidden="true" />
      {message}
    </p>
  );
}

export function BusinessOnboardingForm() {
  const router = useRouter();
  const [phase, setPhase] = useState<Phase>("form");
  const [apiError, setApiError] = useState<string | null>(null);

  const {
    register,
    control,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<OnboardingFormValues, undefined, CreateBusinessInput>({
    resolver: createBusinessResolver,
    defaultValues: {
      logo: undefined,
      name: "",
      description: "",
      type: "service",
      phone: "",
      website: undefined,
      address: "",
      city: "",
      country: "NPL",
    },
    mode: "onChange",
  });

  const onSubmit = async (values: CreateBusinessInput) => {
    setApiError(null);
    setPhase("creating");

    try {
      await createBusiness({
        name: values.name,
        description: values.description,
        logo: values.logo,
        website: values.website,
        phone: values.phone,
        address: values.address,
        city: values.city,
        country: values.country,
        type: values.type,
      });
      setPhase("ready");
    } catch (err) {
      setPhase("form");
      if (isAxiosError(err) && err.response?.data?.message) {
        setApiError(err.response.data.message as string);
      } else {
        setApiError("Something went wrong. Please try again.");
      }
    }
  };

  if (phase !== "form") {
    return (
      <CreatingDashboard phase={phase} onFinish={() => router.push("/")} />
    );
  }

  const typeValue = watch("type");
  const descriptionValue = watch("description");

  return (
    <div className="relative min-h-screen bg-background flex items-center justify-center p-4 overflow-hidden">
      <div className="relative w-full max-w-lg">
        <div className="mb-7 text-center sm:text-left">
          <h1 className="text-[28px] leading-tight font-semibold tracking-tight text-foreground">
            Set up your business
          </h1>
          <p className="text-sm text-muted-foreground mt-1.5">
            Takes about a minute &mdash; you can change everything later.
          </p>
        </div>

        {apiError && (
          <div className="mb-5 flex items-center gap-2.5 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
            <AlertCircle className="w-4 h-4 shrink-0" aria-hidden="true" />
            {apiError}
          </div>
        )}

        <Card className="rounded-2xl border-border/70 p-6 sm:p-8 shadow-[0_1px_2px_rgb(0_0_0/0.04),0_12px_32px_-12px_rgb(0_0_0/0.12)]">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-7">
            <div>
              <p className="text-[11px] font-semibold tracking-wide text-primary uppercase">
                Business profile
              </p>
              <h2 className="text-lg font-semibold tracking-tight mt-0.5">
                Tell us about your business
              </h2>
            </div>

            <div className="space-y-5">
              <div className="flex items-center gap-4 rounded-xl border border-border/70 bg-muted/30 p-4">
                <Controller
                  control={control}
                  name="logo"
                  render={({ field }) => (
                    <ImageUploader
                      maxImages={1}
                      folder="/logos"
                      onUpload={(url) => field.onChange(url)}
                      accept={["image/jpeg", "image/png", "image/webp"]}
                      maxFileSizeBytes={2 * 1024 * 1024}
                    />
                  )}
                />
                <div className="space-y-0.5">
                  <p className="text-sm font-medium">Business logo</p>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    Optional &middot; PNG, JPG, WebP &middot; max 2 MB
                    <br />
                    Square images look best
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                <label
                  htmlFor="name"
                  className="text-sm font-medium leading-none"
                >
                  Business name
                </label>
                <Input
                  id="name"
                  placeholder="Acme Studio"
                  autoFocus
                  className="h-10"
                  {...register("name")}
                />
                <FieldError message={errors.name?.message} />
              </div>

              <div className="space-y-2">
                <div className="flex items-baseline justify-between">
                  <label
                    htmlFor="description"
                    className="text-sm font-medium leading-none"
                  >
                    Description
                  </label>
                  <span className="text-xs text-muted-foreground tabular-nums">
                    {descriptionValue?.length ?? 0}/2000
                  </span>
                </div>
                <Textarea
                  id="description"
                  placeholder="What does your business do? Customers will see this on your page."
                  className="resize-none leading-relaxed"
                  rows={5}
                  {...register("description")}
                />
                <FieldError message={errors.description?.message} />
              </div>
            </div>

            <div className="space-y-4 border-t border-border/60 pt-6">
              <div>
                <h3 className="text-sm font-semibold">Business type</h3>
                <p className="text-sm text-muted-foreground mt-1">
                  This activates the right modules for you.
                </p>
              </div>

              <div
                role="radiogroup"
                aria-label="Business type"
                className="grid gap-3"
              >
                {BUSINESS_TYPES.map((typeOption) => {
                  const Icon = typeOption.icon;
                  const selected = typeValue === typeOption.value;

                  return (
                    <button
                      key={typeOption.value}
                      type="button"
                      role="radio"
                      aria-checked={selected}
                      onClick={() =>
                        setValue("type", typeOption.value, {
                          shouldValidate: true,
                        })
                      }
                      className={cn(
                        "w-full cursor-pointer text-left rounded-xl border p-4",
                        "transition-all duration-200",
                        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 focus-visible:ring-offset-1",
                        selected
                          ? "border-primary/50 bg-primary/[0.04] ring-1 ring-primary/25 shadow-sm"
                          : "border-border bg-transparent hover:border-primary/30 hover:bg-muted/40",
                      )}
                    >
                      <div className="flex items-center gap-3.5">
                        <div
                          className={cn(
                            "w-10 h-10 rounded-lg flex items-center justify-center shrink-0",
                            "transition-colors duration-200",
                            selected
                              ? "bg-primary text-primary-foreground shadow-sm shadow-primary/30"
                              : "bg-muted text-muted-foreground",
                          )}
                        >
                          <Icon
                            className="w-[18px] h-[18px]"
                            aria-hidden="true"
                          />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-semibold">
                            {typeOption.label}
                          </p>
                          <p className="text-xs text-muted-foreground mt-0.5">
                            {typeOption.description}
                          </p>
                        </div>
                        <div
                          className={cn(
                            "w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0",
                            "transition-colors duration-200",
                            selected
                              ? "border-primary bg-primary"
                              : "border-muted-foreground/30",
                          )}
                          aria-hidden="true"
                        >
                          {selected && (
                            <Check className="w-3 h-3 text-primary-foreground motion-safe:animate-in motion-safe:zoom-in-50 duration-150" />
                          )}
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
              <FieldError message={errors.type?.message} />
            </div>

            <div className="space-y-5 border-t border-border/60 pt-6">
              <h3 className="text-sm font-semibold">Contact and location</h3>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label
                    htmlFor="phone"
                    className="text-sm font-medium leading-none"
                  >
                    Phone
                  </label>
                  <Input
                    id="phone"
                    placeholder="+977 98XXXXXXXX"
                    className="h-10"
                    {...register("phone")}
                  />
                  <FieldError message={errors.phone?.message} />
                </div>

                <div className="space-y-2">
                  <div className="flex items-baseline justify-between">
                    <label
                      htmlFor="website"
                      className="text-sm font-medium leading-none"
                    >
                      Website
                    </label>
                    <span className="text-xs text-muted-foreground">
                      optional
                    </span>
                  </div>
                  <Input
                    id="website"
                    placeholder="https://yoursite.com"
                    className="h-10"
                    {...register("website", {
                      setValueAs: optionalUrlValue,
                    })}
                  />
                  <FieldError message={errors.website?.message} />
                </div>
              </div>

              <div className="space-y-2">
                <label
                  htmlFor="address"
                  className="text-sm font-medium leading-none"
                >
                  Address
                </label>
                <Input
                  id="address"
                  placeholder="Thamel, Kathmandu 44600"
                  className="h-10"
                  {...register("address")}
                />
                <FieldError message={errors.address?.message} />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label
                    htmlFor="city"
                    className="text-sm font-medium leading-none"
                  >
                    City
                  </label>
                  <Input
                    id="city"
                    placeholder="Kathmandu"
                    className="h-10"
                    {...register("city")}
                  />
                  <FieldError message={errors.city?.message} />
                </div>

                <div className="space-y-2">
                  <label
                    htmlFor="country"
                    className="text-sm font-medium leading-none"
                  >
                    Country
                  </label>
                  <Controller
                    control={control}
                    name="country"
                    render={({ field }) => (
                      <CountryDropdown
                        defaultValue={field.value}
                        onChange={(country) => field.onChange(country.alpha3)}
                        placeholder="Select country"
                      />
                    )}
                  />
                  <FieldError message={errors.country?.message} />
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end border-t border-border/60 pt-5">
              <Button
                type="submit"
                variant="default"
                className="gap-1.5 h-10 px-5 shadow-sm shadow-primary/25"
              >
                Create business
                <Check className="w-4 h-4" aria-hidden="true" />
              </Button>
            </div>
          </form>
        </Card>

        <p className="text-center text-xs text-muted-foreground mt-5">
          You can update all of this later in Settings
        </p>
      </div>
    </div>
  );
}
